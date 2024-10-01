// cmd/mygrep/main.go
package main

import (
    "fmt"
    "os"
    // "unicode/utf8"
    "github.com/codecrafters-io/grep-starter-go/internal/matcher"
    // "github.com/codecrafters-io/grep-starter-go/internal/io"
	"bufio"
)

func main() {
    if len(os.Args) < 3 || os.Args[1] != "-E" {
        fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
        os.Exit(2)
    }

    pattern := os.Args[2]
    
    scanner := bufio.NewScanner(os.Stdin)
    matchFound := false
    for scanner.Scan() {
        line := scanner.Bytes()
        if matcher.MatchLiteral(line, pattern) {
            fmt.Println(string(line))
            matchFound = true
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
        os.Exit(1)
    }

    if !matchFound {
        os.Exit(1)
    }
}
