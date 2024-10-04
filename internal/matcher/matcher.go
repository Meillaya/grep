// internal/matcher/matcher.go
package matcher

import (
    "bytes"
    "unicode"
)

type Matcher interface {
    Match(line []byte, pattern string) bool
}

type DigitMatcher struct{}
type AlphanumericMatcher struct{}
type PositiveCharGroupMatcher struct {}
type NegativeCharGroupMatcher struct {}

func (pcgm NegativeCharGroupMatcher) Match(line []byte, pattern string) bool {
    if len(pattern) < 4 || pattern[0] != '[' || pattern[1] != '^' || pattern[len(pattern) - 1] != ']' {
        return false
    }

    chars := pattern[2 : len(pattern) - 1]
    for _, r := range string(line) {
        if !bytes.ContainsRune([]byte(chars), r){
            return true
        }
    }
    return false
}   

func (pcgm PositiveCharGroupMatcher) Match(line []byte, pattern string) bool {
    if len(pattern) < 3 || pattern[0] != '[' || pattern[len(pattern)- 1] != ']' {
        return false
    }

    chars := pattern[1 : len(pattern) - 1]
    for _, char := range chars {
        if bytes.ContainsRune(line, char) {
            return true
        }
    }
    return false
}
func (am AlphanumericMatcher) Match(line []byte, pattern string) bool {
    if pattern == "\\w" {
        return containsAlphanumeric(line)
    }
    return bytes.Contains(line, []byte(pattern))
}
func (dm DigitMatcher) Match(line []byte, pattern string) bool {
    if pattern == "\\d" {
        return containsDigit(line)
    }
    return bytes.Contains(line, []byte(pattern))
}

func containsDigit(line []byte) bool {
    for _, r := range string(line) {
        if unicode.IsDigit(r) {
            return true
        }
    }
    return false
}

func containsAlphanumeric(line []byte) bool {
    for _, r := range string(line) {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
            return true
        }
        
    }
    return false
}