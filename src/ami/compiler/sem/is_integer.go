package sem

import "strings"

// isInteger reports whether s parses as a base‑10 integer (allows optional sign).
func isInteger(s string) bool {
    s = strings.TrimSpace(s)
    if s == "" { return false }
    // allow optional sign
    if s[0] == '+' || s[0] == '-' { s = s[1:] }
    if s == "" { return false }
    for i := 0; i < len(s); i++ { if s[i] < '0' || s[i] > '9' { return false } }
    return true
}

