package matcher

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type RegexMatcher struct {
	pattern *regexp.Regexp
}

func NewRegexMatcher(pattern string) (*RegexMatcher, error) {
	processedPattern := preprocessPattern(pattern)
	
	compiledPattern, err := regexp.Compile(processedPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %v", err)
	}
	
	return &RegexMatcher{pattern: compiledPattern}, nil
}

func (rm *RegexMatcher) Match(line []byte, _ string) bool {
	return rm.pattern.Match(line)
}

func preprocessPattern(pattern string) string {
	if utf8.RuneCountInString(pattern) == 0 {
		return pattern
	}

	pattern = strings.ReplaceAll(pattern, "\\d", "[0-9]")
	pattern = strings.ReplaceAll(pattern, "\\w", "[a-zA-Z0-9_]")

	// Handle backreferences
	groupRe := regexp.MustCompile(`\(([^)]*)\)`)
	matches := groupRe.FindAllStringSubmatch(pattern, -1)
	
	if len(matches) != 0 {
		for i, match := range matches {
			pattern = strings.ReplaceAll(pattern, fmt.Sprintf("\\%d", i+1), match[1])
		}
	}

	// Handle nested backreferences
	pattern = handleNestedBackreferences(pattern)

	return pattern
}

func handleNestedBackreferences(pattern string) string {
	stack := []int{}
	result := []byte(pattern)
	
	for i := 0; i < len(result); i++ {
		if result[i] == '(' {
			stack = append(stack, i)
		} else if result[i] == ')' && len(stack) > 0 {
			start := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			
			subPattern := string(result[start+1 : i])
			subPattern = handleNestedBackreferences(subPattern)
			copy(result[start+1:], []byte(subPattern))
		}
	}
	
	return string(result)
}
