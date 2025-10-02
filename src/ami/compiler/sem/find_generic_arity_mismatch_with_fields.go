package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/types"

// findGenericArityMismatchWithFields parses both sides and finds a nested generic arity mismatch.
// Returns a path of generic bases, their argument indices per level, and a struct field path when applicable.
func findGenericArityMismatchWithFields(expected, actual string) (bool, []string, []int, []string, string, int, int) {
    et, eerr := types.Parse(expected)
    at, aerr := types.Parse(actual)
    if eerr != nil || aerr != nil { return false, nil, nil, nil, "", 0, 0 }
    return arityMismatchInTypesWithFields(et, at)
}

