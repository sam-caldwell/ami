package exec

import "unicode"

// SanitizeWorkerSymbol returns a C-safe symbol name by applying:
// - prefixing with `prefix`
// - replacing any non [A-Za-z0-9_] with '_'
// - ensuring the first character after prefix is [A-Za-z_]; if not, prefix 'x_'
// - collapsing consecutive '_' minimally for readability (not required for correctness)
func SanitizeWorkerSymbol(prefix, name string) string {
    // replace invalid runes with '_'
    out := make([]rune, 0, len(prefix)+len(name))
    for _, r := range prefix { out = append(out, r) }
    if name == "" { return string(out) }
    // ensure first is letter or '_'
    first := []rune(name)[0]
    if !(unicode.IsLetter(first) || first == '_') {
        out = append(out, 'x', '_')
    }
    prevUnderscore := false
    for _, r := range name {
        ok := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
        if !ok { r = '_' }
        if r == '_' && prevUnderscore { continue }
        out = append(out, r)
        prevUnderscore = (r == '_')
    }
    return string(out)
}

