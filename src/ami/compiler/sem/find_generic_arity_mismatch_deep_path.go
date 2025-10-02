package sem

// findGenericArityMismatchDeepPath is like findGenericArityMismatchDeep but also returns
// a path of generic base names from the outermost to the mismatching base, and the
// argument indices taken at each nesting level.
func findGenericArityMismatchDeepPath(expected, actual string) (bool, []string, []int, string, int, int) {
    eb, eargs, eok := baseAndArgs(expected)
    ab, aargs, aok := baseAndArgs(actual)
    if !eok || !aok {
        // attempt typed detection (handles Struct)
        if m, path, pathIdx, _, b, w, g := findGenericArityMismatchWithFields(expected, actual); m { return true, path, pathIdx, b, w, g }
        if m, b, w, g := isGenericArityMismatch(expected, actual); m { return true, nil, nil, b, w, g }
        return false, nil, nil, "", 0, 0
    }
    if eb != ab { return false, nil, nil, "", 0, 0 }
    if len(eargs) != len(aargs) { return true, []string{eb}, []int{}, eb, len(eargs), len(aargs) }
    for i := range eargs {
        if m, p, idx, b, w, g := findGenericArityMismatchDeepPath(eargs[i], aargs[i]); m {
            // prepend current base
            return true, append([]string{eb}, p...), append([]int{i}, idx...), b, w, g
        }
    }
    return false, nil, nil, "", 0, 0
}

