package ascii

import "strings"

// wrapTokens wraps a token slice to width w, preserving token boundaries.
func wrapTokens(parts []string, w int) string {
    if w <= 0 { return strings.Join(parts, "") }
    var out []string
    var line strings.Builder
    for _, p := range parts {
        if line.Len()+len(p) > w {
            out = append(out, line.String())
            line.Reset()
        }
        line.WriteString(p)
    }
    if line.Len() > 0 { out = append(out, line.String()) }
    return strings.Join(out, "\n")
}

