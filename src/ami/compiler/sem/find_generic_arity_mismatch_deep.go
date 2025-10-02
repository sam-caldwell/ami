package sem

// findGenericArityMismatchDeep parses expected and actual types and recursively
// searches for a generic arity mismatch at any nesting level. Returns the first
// mismatch encountered (base name and arities). Falls back to top‑level textual
// detection when parsing fails.
func findGenericArityMismatchDeep(expected, actual string) (bool, string, int, int) {
    // Text-based recursive scan that respects nested '<>' and splits top-level commas.
    eb, eargs, eok := baseAndArgs(expected)
    ab, aargs, aok := baseAndArgs(actual)
    if !eok || !aok { return isGenericArityMismatch(expected, actual) }
    if eb != ab { return false, "", 0, 0 }
    if len(eargs) != len(aargs) { return true, eb, len(eargs), len(aargs) }
    // Recurse into paired arguments
    for i := range eargs { if m, b, w, g := findGenericArityMismatchDeep(eargs[i], aargs[i]); m { return m, b, w, g } }
    return false, "", 0, 0
}

