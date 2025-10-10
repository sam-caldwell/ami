package driver

import "strings"

// trimParens removes a single pair of wrapping parentheses when they fully enclose the string.
// It does not alter inner content or nested parentheses depth.
func trimParens(s string) string {
    s = strings.TrimSpace(s)
    if len(s) < 2 { return s }
    if s[0] != '(' || s[len(s)-1] != ')' { return s }
    // ensure outer parens match
    depth := 0
    for i := 0; i < len(s); i++ {
        if s[i] == '(' { depth++ }
        if s[i] == ')' {
            depth--
            if depth == 0 && i != len(s)-1 { return s } // extra trailing tokens
        }
    }
    if depth == 0 { return strings.TrimSpace(s[1:len(s)-1]) }
    return s
}

