package sem

import "strings"

// numericPrefix returns the leading decimal digits of s (trimmed); empty if none.
func numericPrefix(s string) string {
    s = strings.TrimSpace(s)
    i := 0
    for i < len(s) && s[i] >= '0' && s[i] <= '9' { i++ }
    if i == 0 { return "" }
    return s[:i]
}

