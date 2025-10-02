package main

// stringsHasPrefixAny reports whether s has any of the given prefixes.
func stringsHasPrefixAny(s string, prefixes []string) bool {
    for _, p := range prefixes {
        if len(s) >= len(p) && s[:len(p)] == p { return true }
    }
    return false
}

