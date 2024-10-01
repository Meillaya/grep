package matcher

import "bytes"

func MatchLiteral(line []byte, pattern string) bool {
    return bytes.Contains(line, []byte(pattern))
}