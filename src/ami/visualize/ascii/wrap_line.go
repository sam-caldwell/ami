package ascii

import "strings"

// wrapLine wraps s to width w (when w>0), producing a single string possibly containing newlines.
func wrapLine(s string, w int) string {
    if w <= 0 || len(s) <= w { return s }
    var out []string
    for i := 0; i < len(s); i += w {
        j := i + w
        if j > len(s) { j = len(s) }
        out = append(out, s[i:j])
    }
    return strings.Join(out, "\n")
}

