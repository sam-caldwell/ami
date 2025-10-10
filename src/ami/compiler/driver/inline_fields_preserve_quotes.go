package driver

import "strings"

func fieldsPreserveQuotes(s string) []string {
    var out []string
    var cur strings.Builder
    inStr := false
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if ch == '"' { inStr = !inStr; cur.WriteByte(ch); continue }
        if !inStr && (ch == ' ' || ch == '\t') {
            if cur.Len() > 0 { out = append(out, strings.TrimSpace(cur.String())); cur.Reset() }
            continue
        }
        cur.WriteByte(ch)
    }
    if cur.Len() > 0 { out = append(out, strings.TrimSpace(cur.String())) }
    return out
}

