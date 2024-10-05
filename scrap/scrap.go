package matcher

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"unicode"
	"fmt"
)

// RegexMatcher is the main struct for regex matching.
// It contains a memoization cache to store intermediate match results.
type RegexMatcher struct {
	memo map[string]bool // Memoization cache using text and pattern indices as keys
}

// Match determines if the line matches the given pattern.
// It initializes the memoization cache and starts the recursive matching process.
func (rm *RegexMatcher) Match(line []byte, pattern string) bool {
	// Initialize memoization cache
	rm.memo = make(map[string]bool)

	// Validate the pattern for balanced parentheses and brackets
	err := validatePattern(pattern)
	if err != nil {
		// Handle invalid pattern gracefully
		return false
	}

	// Convert pattern to byte slice for efficient access
	patternBytes := []byte(pattern)

	// Start recursive matching from text index 0 and pattern index 0
	return matchRegex(line, patternBytes, [][]byte{}, 0, 0, rm)
}

// validatePattern checks for balanced parentheses and brackets in the regex pattern.
// It returns an error if the pattern is malformed.
func validatePattern(pattern string) error {
	var stack []rune
	for _, ch := range pattern {
		switch ch {
		case '(', '[':
			stack = append(stack, ch)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return errors.New("unbalanced parentheses")
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return errors.New("unbalanced brackets")
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) != 0 {
		return errors.New("unbalanced pattern")
	}
	return nil
}

// matchRegex recursively matches the text against the pattern with captures and memoization.
// text: the input text
// pattern: the regex pattern
// captures: a slice of byte slices holding captured groups
// textIdx: current index in text
// patIdx: current index in pattern
// rm: pointer to RegexMatcher containing the memoization cache
func matchRegex(text []byte, pattern []byte, captures [][]byte, textIdx, patIdx int, rm *RegexMatcher) bool {
	// Create a unique key for memoization based on current textIdx and patIdx
	key := strconv.Itoa(textIdx) + "|" + strconv.Itoa(patIdx) + "|" + capturesKey(captures)
	if val, exists := rm.memo[key]; exists {
		return val
	}

	// Base case: If pattern is fully matched, text must also be fully matched
	if patIdx == len(pattern) {
		rm.memo[key] = textIdx == len(text)
		return textIdx == len(text)
	}

	// Initialize match result to false
	rm.memo[key] = false

	// Handle end anchor '$'
	if pattern[patIdx] == '$' {
		if textIdx == len(text) {
			rm.memo[key] = true
			return true
		}
		return false
	}

	// Handle start anchor '^'
	if pattern[patIdx] == '^' {
		match := matchRegex(text, pattern, captures, 0, patIdx+1, rm)
		rm.memo[key] = match
		return match
	}

	// Handle backreferences and escape sequences
	if pattern[patIdx] == '\\' {
		if patIdx+1 >= len(pattern) {
			// Invalid escape sequence
			return false
		}
		escapedChar := pattern[patIdx+1]
		switch escapedChar {
		case 'w':
			// \w matches word characters [a-zA-Z0-9_]
			if textIdx >= len(text) || !isWord(rune(text[textIdx])) {
				return false
			}
			return processAtom(text, pattern, captures, textIdx, patIdx, rm)
		case 'd':
			// \d matches digits [0-9]
			if textIdx >= len(text) || !unicode.IsDigit(rune(text[textIdx])) {
				return false
			}
			return processAtom(text, pattern, captures, textIdx, patIdx, rm)
		case 's':
			// \s matches whitespace characters
			if textIdx >= len(text) || !unicode.IsSpace(rune(text[textIdx])) {
				return false
			}
			return processAtom(text, pattern, captures, textIdx, patIdx, rm)
		default:
			// Handle backreferences like \1, \2, etc.
			if unicode.IsDigit(rune(escapedChar)) {
				captureIndex, _ := strconv.Atoi(string(escapedChar))
				if captureIndex <= 0 || captureIndex > len(captures) {
					return false
				}
				capture := captures[captureIndex-1]
				if textIdx+len(capture) > len(text) {
					return false
				}
				if !bytes.Equal(text[textIdx:textIdx+len(capture)], capture) {
					return false
				}
				// Recursive call after matching backreference
				return matchRegex(text, pattern, captures, textIdx+len(capture), patIdx+2, rm)
			}
			// Unsupported escape sequence
			return false
		}
	}

	// Handle character classes [abc] or [^abc]
	if pattern[patIdx] == '[' {
		endIdx := findClosingBrace(pattern, patIdx)
		if endIdx == -1 {
			// Malformed character class
			return false
		}
		isNegated := false
		start := patIdx + 1
		if pattern[start] == '^' {
			isNegated = true
			start++
		}
		charClass := pattern[start:endIdx]
		if textIdx >= len(text) {
			return false
		}
		if matchChar(text[textIdx], charClass, isNegated) {
			return processAtom(text, pattern, captures, textIdx, patIdx, rm)
		}
		return false
	}

	// Handle wildcard '.'
	if pattern[patIdx] == '.' {
		if textIdx >= len(text) {
			return false
		}
		return processAtom(text, pattern, captures, textIdx, patIdx, rm)
	}

	// Handle literal character
	if textIdx < len(text) && pattern[patIdx] == text[textIdx] {
		return processAtom(text, pattern, captures, textIdx, patIdx, rm)
	}

	// No match
	return false
}

// processAtom processes the current atom and handles any quantifiers that follow it.
// It returns true if a match is found, false otherwise.
// text, pattern: input text and pattern
// captures: current captures
// textIdx, patIdx: current indices in text and pattern
// rm: pointer to RegexMatcher containing the memoization cache
func processAtom(text []byte, pattern []byte, captures [][]byte, textIdx, patIdx int, rm *RegexMatcher) bool {
	// Determine if the next character in pattern is a quantifier
	nextPatIdx := patIdx + 1
	if nextPatIdx >= len(pattern) {
		// No quantifier, proceed to match the next characters
		return matchRegex(text, pattern, captures, textIdx+1, patIdx+1, rm)
	}

	nextChar := pattern[nextPatIdx]
	var quantifier string
	var min, max int
	var isNonGreedy bool

	// Check for non-greedy quantifiers like *?, +?, ??, {n,m}?
	switch {
	case nextChar == '?' && (patIdx > 0 && (pattern[patIdx-1] == '*' || pattern[patIdx-1] == '+' || pattern[patIdx-1] == '?' || pattern[patIdx-1] == '}')):
		isNonGreedy = true
		quantifier = "non-greedy"
	case nextChar == '*' || nextChar == '+' || nextChar == '?':
		quantifier = string(nextChar)
		if nextPatIdx+1 < len(pattern) && pattern[nextPatIdx+1] == '?' {
			isNonGreedy = true
			quantifier += "?"
		}
	case nextChar == '{':
		endBrace := findClosingBrace(pattern, nextPatIdx)
		if endBrace == -1 {
			return false
		}
		quantText := string(pattern[nextPatIdx+1 : endBrace])
		if strings.Contains(quantText, ",") {
			parts := strings.Split(quantText, ",")
			var err error
			min, err = strconv.Atoi(parts[0])
			if err != nil {
				return false
			}
			if parts[1] == "" {
				max = -1 // No upper limit
			} else {
				max, err = strconv.Atoi(parts[1])
				if err != nil {
					return false
				}
			}
		} else {
			var err error
			min, err = strconv.Atoi(quantText)
			if err != nil {
				return false
			}
			max = min
		}
		if endBrace+1 < len(pattern) && pattern[endBrace+1] == '?' {
			isNonGreedy = true
			endBrace++
		}
		quantifier = "{" + quantText + "}"
		patIdx = endBrace // Update patIdx to skip the quantifier
	}
	fmt.Println("Max value:", max)
	if quantifier == "" || quantifier == "non-greedy" {
		// No quantifier or only indicating non-greedy without a specific quantifier
		return matchRegex(text, pattern, captures, textIdx+1, patIdx+1, rm)
	}

	switch quantifier {
	case "*":
		// Match zero or more of the current atom
		return handleQuantifier(text, pattern, captures, textIdx, patIdx, 0, -1, isNonGreedy, rm)
	case "+", "+?":
		// Match one or more of the current atom
		return handleQuantifier(text, pattern, captures, textIdx, patIdx, 1, -1, isNonGreedy, rm)
	case "?", "??":
		// Match zero or one of the current atom
		return handleQuantifier(text, pattern, captures, textIdx, patIdx, 0, 1, isNonGreedy, rm)
	default:
		if strings.HasPrefix(quantifier, "{") && strings.HasSuffix(quantifier, "}") {
			quantContent := quantifier[1 : len(quantifier)-1]
			min, max, err := parseQuantifier(quantContent)
			if err != nil {
				return false
			}
			return handleQuantifier(text, pattern, captures, textIdx, patIdx, min, max, isNonGreedy, rm)
		}
		// Unsupported quantifier
		return false
	}
}


func parseQuantifier(quantContent string) (min, max int, err error) {
    parts := strings.Split(quantContent, ",")
    if len(parts) == 1 {
        min, err = strconv.Atoi(parts[0])
        max = min
    } else if len(parts) == 2 {
        min, err = strconv.Atoi(parts[0])
        if err != nil {
            return 0, 0, err
        }
        if parts[1] == "" {
            max = -1
        } else {
            max, err = strconv.Atoi(parts[1])
        }
    } else {
        err = errors.New("invalid quantifier format")
    }
    return min, max, err
}
// handleQuantifier processes quantifiers (*, +, ?, {n}, {n,m}).
// min: minimum number of matches
// max: maximum number of matches (-1 for unlimited)
// isNonGreedy: whether the quantifier is non-greedy
// It returns true if a match is found, false otherwise.
func handleQuantifier(text []byte, pattern []byte, captures [][]byte, textIdx, patIdx, min, max int, isNonGreedy bool, rm *RegexMatcher) bool {
	// count := 0
	startIdx := textIdx
	currentPatIdx := patIdx

	// Determine the type of atom to match
	// For simplicity, assume single-byte atoms; extend as needed for multi-byte
	// var atom byte
	// if pattern[currentPatIdx] == '.' || pattern[currentPatIdx] == '[' || pattern[currentPatIdx] == '(' || pattern[currentPatIdx] == '\\' {
	// 	// These are handled in the main matching function
	// 	atom = pattern[currentPatIdx]
	// } else {
	// 	atom = pattern[currentPatIdx]
	// }

	// Collect possible match lengths based on quantifier constraints
	matchLengths := []int{}

	if isNonGreedy {
		// Non-greedy: attempt smallest number of matches first
		for i := min; (max == -1 || i <= max) && textIdx+i <= len(text); i++ {
			if i > 0 {
				if !matchAtoms(text, currentPatIdx, textIdx, i, pattern) {
					break
				}
			}
			matchLengths = append(matchLengths, i)
		}
	} else {
		// Greedy: attempt largest number of matches first
		var maxMatchRange int
		if max == -1 {
			maxMatchRange = len(text) - textIdx
		} else {
			if max > len(text)-textIdx {
				maxMatchRange = len(text) - textIdx
			} else {
				maxMatchRange = max
			}
		}
		for i := maxMatchRange; i >= min; i-- {
			if i > 0 {
				if !matchAtoms(text, currentPatIdx, textIdx, i, pattern) {
					continue
				}
			}
			matchLengths = append(matchLengths, i)
		}
	}

	// Attempt to match for each possible count
	for _, i := range matchLengths {
		newTextIdx := startIdx + i
		match := matchRegex(text, pattern, captures, newTextIdx, patIdx+2, rm)
		if match {
			return true
		}
	}

	return false
}

// matchAtoms checks if a substring of length 'length' starting at 'textIdx' matches the atom at 'patIdx'.
// It returns true if all atoms match, false otherwise.
func matchAtoms(text []byte, patIdx, textIdx, length int, pattern []byte) bool {
	for i := 0; i < length; i++ {
		currentTextIdx := textIdx + i
		if currentTextIdx >= len(text) {
			return false
		}
		if !matchesAtom(text[currentTextIdx], pattern, patIdx) {
			return false
		}
	}
	return true
}

// matchesAtom checks if a single character matches the regex atom.
// It returns true if it matches, false otherwise.
func matchesAtom(c byte, pattern []byte, patIdx int) bool {
	switch pattern[patIdx] {
	case '.':
		return true
	case '\\':
		if patIdx+1 >= len(pattern) {
			return false
		}
		escapedChar := pattern[patIdx+1]
		switch escapedChar {
		case 'w':
			return isWord(rune(c))
		case 'd':
			return unicode.IsDigit(rune(c))
		case 's':
			return unicode.IsSpace(rune(c))
		default:
			// Unsupported escape sequence
			return false
		}
	default:
		return c == pattern[patIdx]
	}
}

// handleCapturingGroup handles capturing groups by storing the captured text.
func handleCapturingGroup(text []byte, pattern []byte, captures [][]byte, textIdx, patIdx int, rm *RegexMatcher) bool {
	closingParen := findClosingParen(pattern, patIdx)
	if closingParen == -1 {
		// Malformed pattern
		return false
	}

	subPattern := pattern[patIdx+1 : closingParen]
	// Attempt to capture all possible substrings
	for i := 0; i <= len(text)-textIdx; i++ {
		if matchRegex(text, subPattern, captures, textIdx, patIdx+1, rm) {
			newCaptures := append(captures, text[textIdx:textIdx+i])
			if matchRegex(text, pattern, newCaptures, textIdx+i, closingParen+1, rm) {
				return true
			}
		}
	}
	return false
}

// splitPattern splits the pattern by unescaped '|' characters, considering group nesting.
func splitPattern(pattern string) []string {
	var parts []string
	var current bytes.Buffer
	escaped := false
	level := 0
	for _, ch := range pattern {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			current.WriteRune(ch)
			continue
		}
		if ch == '(' || ch == '[' {
			level++
		}
		if ch == ')' || ch == ']' {
			if level > 0 {
				level--
			}
		}
		if ch == '|' && level == 0 {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(ch)
	}
	parts = append(parts, current.String())
	return parts
}

// findClosingParen finds the index of the matching closing parenthesis for an opening parenthesis at patIdx.
func findClosingParen(pattern []byte, patIdx int) int {
	if patIdx >= len(pattern) || pattern[patIdx] != '(' {
		return -1
	}
	count := 0
	for i := patIdx; i < len(pattern); i++ {
		if pattern[i] == '(' {
			count++
		} else if pattern[i] == ')' {
			count--
			if count == 0 {
				return i
			}
		}
	}
	return -1
}

// findClosingBrace finds the index of the matching closing brace for an opening brace at patIdx.
func findClosingBrace(pattern []byte, patIdx int) int {
	if patIdx >= len(pattern) || pattern[patIdx] != '{' {
		return -1
	}
	for i := patIdx + 1; i < len(pattern); i++ {
		if pattern[i] == '}' {
			return i
		}
	}
	return -1
}

// isWord checks if a rune is a word character.
func isWord(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// matchChar checks if a character matches the character class.
// chars: characters inside the []
// isNegated: true if the character class is negated [^]
func matchChar(c byte, chars []byte, isNegated bool) bool {
	contains := bytes.Contains(chars, []byte{c})
	if isNegated {
		return !contains
	}
	return contains
}

// capturesKey generates a unique key for the current captures to use in memoization.
func capturesKey(captures [][]byte) string {
	var parts []string
	for _, cap := range captures {
		parts = append(parts, string(cap))
	}
	return strings.Join(parts, "|")
}
