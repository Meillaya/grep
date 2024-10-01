// internal/matcher/matcher.go
package matcher

type Matcher interface {
    Match(line []byte, pattern string) bool
}