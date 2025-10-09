package main

// sanitizeForCSymbol mirrors runtime sanitize for C symbol creation in builder.
func sanitizeForCSymbol(prefix, name string) string {
    out := []rune(prefix)
    if name == "" { return string(out) }
    r := []rune(name)
    first := r[0]
    if !((first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z') || first == '_') {
        out = append(out, 'x', '_')
    }
    prevUnderscore := false
    for _, ch := range r {
        ok := (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_'
        if !ok { ch = '_' }
        if ch == '_' && prevUnderscore { continue }
        out = append(out, ch)
        prevUnderscore = (ch == '_')
    }
    return string(out)
}

