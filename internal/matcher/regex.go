package matcher

import (
	"bytes"
	"unicode"
)

type RegexMatcher struct {}		


func (rm RegexMatcher) Match(line []byte, pattern string) bool {
	return matchRegex(line, []byte(pattern))
}

func matchRegex(text, pattern []byte) bool {
	if len(pattern) == 0 {
		return true
	}

	if len(text) == 0 {
		return false
	}

	switch pattern[0] {
	case '\\':
		if len(pattern) > 1 {
			switch pattern[1] {
				case 'd':
					if unicode.IsDigit(rune(text[0])) {
						return matchRegex(text[1:], pattern[2:])
					}
				case 'w':
					if isWord(rune(text[0])) {
						return matchRegex(text[1:], pattern[2:])
					}
				}
			}
		case '[':
			if len(pattern) > 1 {
				if pattern[1] == '^' {
					return matchNegativeCharGroup(text, pattern)
				}
				return matchPositveCharGroup(text, pattern)
			}
		case '.':
			return matchRegex(text[1:], pattern[1:])
		default:
			if text[0] == pattern[0] {
				return matchRegex(text[1:], pattern[1:])
			}
		}

		return matchRegex(text[1:], pattern)
}
func matchPositveCharGroup(text []byte, pattern []byte) bool {

	end := bytes.IndexByte(pattern, ']')
	if end == - 1 {
		// malformed pattern
		return false 
	}

	chars := pattern[1:end]

	if bytes.ContainsAny(text[:1], string(chars)) {
		return matchRegex(text[1:], pattern[end + 1:])
	}

	return matchRegex(text[1:], pattern)
}

func matchNegativeCharGroup(text []byte, pattern []byte) bool {
	
	end := bytes.IndexByte(pattern, ']')

	if end == - 1 {
		//malformed pattern
		return false
	}

	chars := pattern[2:end]

	if !bytes.ContainsAny(text[:1], string(chars)) {
		return matchRegex(text[1:], pattern[end + 1:])
	}

	return matchRegex(text[1:], pattern)
	
}
func isWord(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}