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