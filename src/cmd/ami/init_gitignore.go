package main

import (
    "os"
    "strings"
)

// appendLineIfMissing adds a line to a file if it is not already present.
func appendLineIfMissing(path string, line string) {
    data, _ := os.ReadFile(path)
    if !strings.Contains(string(data), line) {
        f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
        if err != nil {
            return
        }
        defer f.Close()
        _, _ = f.WriteString(line)
    }
}

