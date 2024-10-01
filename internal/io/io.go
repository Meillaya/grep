// internal/io/io.go
package io

import (
    "bufio"
    "os"
)

func ReadLines(reader *os.File) (*bufio.Scanner, error) {
    return bufio.NewScanner(reader), nil
}