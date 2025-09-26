package root

import "strings"

func splitIdents(s string) []string {
    out := []string{}
    cur := strings.Builder{}
    flush := func() {
        if cur.Len() > 0 {
            out = append(out, cur.String())
            cur.Reset()
        }
    }
    for _, r := range s {
        if r == '-' || r == '.' || r == ':' || r == '/' || r == '\\' || r == '*' || r == '[' || r == ']' || r == '<' || r == '>' || r == '(' || r == ')' || r == ',' || r == ';' || r == '|' || r == '!' || r == '?' || r == '=' || r == '+' || r == '&' || r == '^' || r == '%' || r == '$' || r == '#' || r == '@' || r == '~' || r == '`' || r == ' ' || r == '\t' || r == '\n' || r == '\r' {
            flush()
            continue
        }
        cur.WriteRune(r)
    }
    flush()
    return out
}
