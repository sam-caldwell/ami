package sem

// splitTopLevel splits a comma-separated list, ignoring commas inside '<>' pairs.
func splitTopLevel(s string) []string {
    var out []string
    depth := 0
    start := 0
    for i, r := range s {
        switch r {
        case '<': depth++
        case '>': if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                out = append(out, s[start:i])
                start = i + 1
            }
        }
    }
    if start <= len(s) { out = append(out, s[start:]) }
    return out
}

