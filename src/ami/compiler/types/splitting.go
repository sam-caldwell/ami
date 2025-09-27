package types

import "strings"

// splitAllTop splits by commas at the top level of a generic argument list,
// without breaking nested <...> groups.
func splitAllTop(s string) []string {
    var out []string
    depth := 0
    start := 0
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case '<':
            depth++
        case '>':
            if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                out = append(out, strings.TrimSpace(s[start:i]))
                start = i + 1
            }
        }
    }
    out = append(out, strings.TrimSpace(s[start:]))
    return out
}

