package matcher

import (
	"strings"
	"testing"
	"time"

	"github.com/codecrafters-io/grep-starter-go/internal/matcher"
)

func TestRegexMatcher(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected bool
	}{
		// Basic Matching
		{"Literal match", "hello", "hello", true},
		{"Literal mismatch", "hello", "hella", false},
		{"Dot wildcard match", "h3llo", "h.llo", true},
		{"Dot wildcard mismatch", "hlllo", "h.llo", false},

		// Character Classes
		{"Positive char class match", "a", "[abc]", true},
		{"Positive char class mismatch", "d", "[abc]", false},
		{"Negative char class match", "d", "[^abc]", true},
		{"Negative char class mismatch", "a", "[^abc]", false},

		// Quantifiers
		{"Zero or more quantifier match", "aaa", "a*", true},
		{"One or more quantifier match", "aaa", "a+", true},
		{"Zero or one quantifier match", "a", "a?", true},
		{"Exact count quantifier match", "aaa", "a{3}", true},
		{"Range quantifier match", "aaaabbbb", "a{2,4}b{2,4}", true},

		// Non-Greedy Quantifiers
		{"Non-greedy quantifier match", "aaab", "a+?b", true},

		// Anchors
		{"Start anchor match", "abcde", "^abc", true},
		{"End anchor match", "abcde", "cde$", true},

		// Capturing Groups and Backreferences
		{"Simple backreference match", "cat and cat", "(cat) and \\1", true},
		{"Complex backreference match", "grep 101 is doing grep 101 times", "(\\w{4} \\d{3}) is doing \\1 times", true},

		// Alternation
		{"Alternation match first option", "apple", "a|b", true},
		{"Alternation match second option", "banana", "a|b", true},

		// Nested Groups and Quantifiers
		{"Nested group with quantifier match", "ababab", "(ab)*", true},
		{"Non-capturing group with quantifier", "abcdcdcd", "(?:ab|cd)+", true},

		// Multiple and Nested Backreferences
		{"Multiple backreferences match", "abcabc", "(abc)\\1", true},
		{"Nested backreferences match", "abab", "((ab))\\1", true},
	}

	for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            rm, err := matcher.NewRegexMatcher(tc.pattern)
            if err != nil {
                t.Fatalf("Failed to create RegexMatcher: %v", err)
            }
            result := rm.Match([]byte(tc.text), tc.pattern)
            if result != tc.expected {
                t.Errorf("Pattern: '%s' Text: '%s' Expected: %v, Got: %v",
                    tc.pattern, tc.text, tc.expected, result)
            }
        })
    }
}



func TestAdvancedRegexPatterns(t *testing.T) {
    tests := []struct {
        name     string
        text     string
        pattern  string
        expected bool
    }{
        {"Complex alternation", "abcdefg", "a(b|c|d){3}g", true},
        {"Nested quantifiers", "aaaabbbbbcccccc", "(a+b+c+){1,2}", true},
        {"Lookahead", "hello world", "hello(?=\\sworld)", true},
        {"Negative lookahead", "hello universe", "hello(?!\\sworld)", true},
        {"Word boundaries", "cat in the hat", "\\bcat\\b.*\\bhat\\b", true},
        {"Non-word boundaries", "helloworld", "hello\\Bworld", true},
        {"Backreference with quantifier", "catcatcat", "(cat)\\1+", true},
        {"Complex character class", "a1B2c3D4", "[a-z][0-9][A-Z][0-9][a-z][0-9]", true},
        {"Negated character class", "A1b2C3", "[^a-z][^A-Z][^0-9]{2}", true},
        {"Unicode support", "こんにちは世界", "\\p{Hiragana}+\\p{Han}+", true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            rm, err := matcher.NewRegexMatcher(tc.pattern)
            if err != nil {
                t.Fatalf("Failed to create RegexMatcher: %v", err)
            }
            result := rm.Match([]byte(tc.text), tc.pattern)
            if result != tc.expected {
                t.Errorf("Pattern: '%s' Text: '%s' Expected: %v, Got: %v",
                    tc.pattern, tc.text, tc.expected, result)
            }
        })
    }
}

func TestRegexPerformance(t *testing.T) {
    longText := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 1000)
    complexPattern := "a.*z"

    rm, err := matcher.NewRegexMatcher(complexPattern)
    if err != nil {
        t.Fatalf("Failed to create RegexMatcher: %v", err)
    }

    start := time.Now()
    result := rm.Match([]byte(longText), complexPattern)
    duration := time.Since(start)

    t.Logf("Matching time: %v", duration)
    if !result {
        t.Errorf("Expected match for long text, but got no match")
    }

    if duration > time.Second {
        t.Errorf("Matching took too long: %v", duration)
    }
}

func TestRegexEdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        text     string
        pattern  string
        expected bool
    }{
        {"Empty text and pattern", "", "", true},
        {"Empty text with non-empty pattern", "", "a*", true},
        {"Non-empty text with empty pattern", "abc", "", true},
        {"Pattern with only quantifiers", "aaa", "*+?", false},
        {"Unmatched parentheses", "abc", "(abc", false},
        {"Unmatched brackets", "abc", "[abc", false},
        {"Invalid backreference", "abc", "\\1(a)", false},
        {"Nested capturing groups", "abcabc", "((a)(b)(c))\\1", true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            rm, err := matcher.NewRegexMatcher(tc.pattern)
            if err != nil {
                // For invalid patterns, we expect an error
                if tc.expected {
                    t.Fatalf("Failed to create RegexMatcher: %v", err)
                }
                return
            }
            result := rm.Match([]byte(tc.text), tc.pattern)
            if result != tc.expected {
                t.Errorf("Pattern: '%s' Text: '%s' Expected: %v, Got: %v",
                    tc.pattern, tc.text, tc.expected, result)
            }
        })
    }
}

func TestRegexMatcherInvalidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{"Unclosed parenthesis", "(abc"},
		{"Unclosed bracket", "[abc"},
		{"Unclosed brace", "a{2,"},
		{"Invalid quantifier", "a{,}"},
		{"Invalid backreference", "\\2(a)"},
	}

	for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            _, err := matcher.NewRegexMatcher(tc.pattern)
            if err == nil {
                t.Errorf("Expected an error for invalid pattern '%s', but got nil", tc.pattern)
            }
        })
    }
}
