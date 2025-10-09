package driver

import "strings"

// atoiSafe parses a positive integer prefix of s.
// Returns (value, true) on success; (0, false) otherwise.
func atoiSafe(s string) (int, bool) {
    s = strings.TrimSpace(s)
    if s == "" { return 0, false }
    n := 0
    ok := false
    for i := 0; i < len(s); i++ {
        c := s[i]
        if c < '0' || c > '9' { break }
        ok = true
        n = n*10 + int(c-'0')
        if n < 0 { return 0, false }
    }
    return n, ok
}
