package main

import "strings"

// splitLinesPreserve splits into lines, preserving empty trailing line behavior.
func splitLinesPreserve(s string) []string {
    if s == "" { return []string{""} }
    s = strings.ReplaceAll(s, "\r\n", "\n")
    s = strings.ReplaceAll(s, "\r", "\n")
    return strings.Split(s, "\n")
}

