package sem

import (
    "sort"
)

// findGenericArityMismatchDeepPathTextFields attempts to discover a generic arity
// mismatch and return both generic path/pathIdx and a textual struct fieldPath
// without relying on types.Parse. It handles simple Struct{...} and Optional<...>
// text forms and delegates generic path detection to findGenericArityMismatchDeepPath
// when not within a struct field.
func findGenericArityMismatchDeepPathTextFields(expected, actual string) (bool, []string, []int, []string, string, int, int) {
    // Optional wrappers
    if be, argsE, okE := baseAndArgs(expected); okE && be == "Optional" && len(argsE) == 1 {
        if ba, argsA, okA := baseAndArgs(actual); okA && ba == "Optional" && len(argsA) == 1 {
            return findGenericArityMismatchDeepPathTextFields(argsE[0], argsA[0])
        }
    }
    // Struct traversal
    if isStructText(expected) && isStructText(actual) {
        ef, okE := parseStructFieldsText(expected)
        af, okA := parseStructFieldsText(actual)
        if okE && okA {
            // common fields in stable order
            keys := make([]string, 0)
            for k := range ef { if _, ok := af[k]; ok { keys = append(keys, k) } }
            sort.Strings(keys)
            for _, k := range keys {
                einner := ef[k]
                ainner := af[k]
                if m, p, idx, fp, b, w, g := findGenericArityMismatchDeepPathTextFields(einner, ainner); m {
                    // prepend field name only (normalized)
                    return true, p, idx, append([]string{k}, fp...), b, w, g
                }
                if m2, p2, idx2, b2, w2, g2 := findGenericArityMismatchDeepPath(einner, ainner); m2 {
                    return true, p2, idx2, []string{k}, b2, w2, g2
                }
            }
        }
    }
    // Fallback to generic deep path detection without fieldPath
    if m, p, idx, b, w, g := findGenericArityMismatchDeepPath(expected, actual); m { return true, p, idx, nil, b, w, g }
    return false, nil, nil, nil, "", 0, 0
}
