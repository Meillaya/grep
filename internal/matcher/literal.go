package matcher

import (
	"bytes"
    
)

type LiteralMatcher struct{}

func (lm LiteralMatcher) Match(line []byte, pattern string) bool {
    return bytes.Contains(line, []byte(pattern))
}