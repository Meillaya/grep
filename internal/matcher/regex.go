package matcher

import (
	"bytes"
	"strings"
	"unicode"
)

type RegexMatcher struct {}		


func (rm RegexMatcher) Match(line []byte, pattern string) bool {

	if strings.HasPrefix(pattern, "^") {
		// If the pattern starts with ^, it must match at the beginning of the line
        return bytes.HasPrefix(line, []byte(pattern[1:]))

	} else if strings.HasSuffix(pattern, "$") {

		// If the pattern starts with $, it must match at the end of the line
		return bytes.HasSuffix(line, []byte(strings.TrimSuffix(pattern, "$")))
		
	} else if strings.Contains(pattern, "(") && strings.Contains(pattern, "|") {

		return matchAlternation(line, pattern)

	}  else if strings.Contains(pattern, "+") || strings.Contains(pattern, "?") {
		return matchRegex(line, []byte(pattern))
	}

	//For patterns without ^, $, + or ?, use existing logic
	return matchRegex(line, []byte(pattern))
}

func matchRegex(text, pattern []byte) bool {
	if len(pattern) == 0 {
		return true
	}

	if len(text) == 0 {
		return false
	}

	if pattern[0] == '^' {
		return matchStartOfLine(text, pattern[1:])
	}

	if len(pattern) > 1 && pattern[1] == '?' {
		if text[0] == pattern[0] {
			return matchRegex(text[1:], pattern[2:]) || matchRegex(text, pattern[2:])
		}
		return matchRegex(text, pattern[2:])
	}


	switch pattern[0] {
	case '$':
		//We should be at the end of the text
		return len(pattern) == 1 && len(text) == 0	
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
			if len(pattern) > 1 && pattern[1] == '+'{
				if text[0] != pattern[0] {
					return false
				}
				i := 1
				for i < len(text) && text[i] == pattern[0] {
					i++
				}
				return matchRegex(text[i:], pattern[2:])
			}else if text[0] == pattern[0] {
				return matchRegex(text[1:], pattern[1:])
			}
		}

		if text[0] == pattern[0] || pattern[0] == '.' {
			return matchRegex(text[1:], pattern[1:])
		}
		
		return matchRegex(text[1:], pattern)
}

func matchAlternation(text []byte, pattern string) bool {

	parts := strings.SplitN(pattern, "(", 2)
	prefix := parts[0]

	if len (parts) < 2 {
		return bytes.Contains(text, []byte(pattern))
	}

	alternatives := strings.Split(strings.TrimSuffix(parts[1], ")"), "|")

	for _, alt := range alternatives {
		fullPattern := prefix + alt
		if bytes.Contains(text, []byte(fullPattern)) {
			return true
		}
	}
	return false
}
func matchStartOfLine(text, pattern []byte) bool {

	return matchRegex(text, pattern)
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