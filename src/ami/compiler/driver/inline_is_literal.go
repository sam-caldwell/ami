package driver

import "strings"

func isLiteral(s string) bool {
    s = strings.TrimSpace(s)
    if s == "" { return false }
    if s == "true" || s == "false" { return true }
    if s[0] == '"' && s[len(s)-1] == '"' { return true }
    dot := false
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if ch >= '0' && ch <= '9' { continue }
        if ch == '.' && !dot { dot = true; continue }
        if i == 0 && (ch == '+' || ch == '-') { continue }
        return false
    }
    return true
}

