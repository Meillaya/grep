// cmd/mygrep/main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/grep-starter-go/internal/matcher"
)

func main() {
    if len(os.Args) < 3 || os.Args[1] != "-E" {
        fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
        os.Exit(2)
    }

    pattern := os.Args[2]
    
    scanner := bufio.NewScanner(os.Stdin)
    var m matcher.Matcher
    
    if strings.HasPrefix(pattern, "[^") && strings.HasSuffix(pattern, "]") {
        m = matcher.NegativeCharGroupMatcher{}
    } else if strings.HasPrefix(pattern, "[") && strings.HasSuffix(pattern, "]") {
        m = matcher.PositiveCharGroupMatcher{}
    } else if pattern == "\\w" {
        m = matcher.AlphanumericMatcher{}
    } else if pattern == "\\d" {
        m = matcher.DigitMatcher{}
    } else {
        m = matcher.LiteralMatcher{}
    }


    matchFound := false
    for scanner.Scan() {
        line := scanner.Bytes()
        if m.Match(line, pattern) {
            matchFound = true
            break
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
        os.Exit(1)
    }

    if !matchFound {
        os.Exit(1)
    } else {
        os.Exit(0)
    }
    
}
