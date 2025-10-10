package driver

import "strings"

// extractInlineBody returns the string inside the outermost braces of a func literal.
func extractInlineBody(s string) string {
    i := strings.Index(s, "{")
    if i < 0 { return "" }
    depth := 0
    for idx := i; idx < len(s); idx++ {
        if s[idx] == '{' { depth++ }
        if s[idx] == '}' {
            depth--
            if depth == 0 {
                inner := s[i+1 : idx]
                return strings.TrimSpace(inner)
            }
        }
    }
    return ""
}

