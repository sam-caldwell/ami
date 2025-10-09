package driver

import "strings"

// parseIntList3 parses a bracketed list like "[a,b,c]" into up to 3 ints.
// Missing entries default to zero. Returns ok=false on malformed input.
func parseIntList3(s string) ([3]int, bool) {
    var out [3]int
    s = strings.TrimSpace(s)
    if s == "" { return out, false }
    // Trim quotes if present
    if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
        s = s[1:len(s)-1]
    }
    s = strings.TrimSpace(s)
    if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
        return out, false
    }
    inner := strings.TrimSpace(s[1 : len(s)-1])
    if inner == "" { return out, true }
    parts := strings.Split(inner, ",")
    n := 0
    for i := 0; i < len(parts) && i < 3; i++ {
        v, ok := atoiSafe(strings.TrimSpace(parts[i]))
        if !ok { return out, false }
        out[i] = v
        n++
    }
    if n == 0 { return out, false }
    return out, true
}

