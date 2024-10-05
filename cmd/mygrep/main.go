// cmd/mygrep/main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"github.com/codecrafters-io/grep-starter-go/internal/matcher"
)

func main() {
    if len(os.Args) < 3 || os.Args[1] != "-E" {
        fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
        os.Exit(2)
    }

    pattern := os.Args[2]
    
    scanner := bufio.NewScanner(os.Stdin)

    regexMatcher, err := matcher.NewRegexMatcher(pattern)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error compiling regex: %v\n", err)
        os.Exit(1)
    }
    matchFound := false

    for scanner.Scan() {
        line := scanner.Text()
        if regexMatcher.Match([]byte(line), line) {
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