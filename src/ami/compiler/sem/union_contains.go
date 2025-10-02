package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/types"

// unionContains returns true if actual type is included in expected union.
// If both are unions, require each alternative of actual to be included in expected.
func unionContains(expected types.Type, actual types.Type) bool {
    eu, eok := expected.(types.Union)
    if !eok {
        // fallback: equality or generic name match
        return types.Equal(expected, actual)
    }
    // helper to test membership
    in := func(t types.Type) bool {
        for _, a := range eu.Alts { if types.Equal(a, t) { return true } }
        return false
    }
    if au, ok := actual.(types.Union); ok {
        for _, t := range au.Alts { if !in(t) { return false } }
        return true
    }
    return in(actual)
}

