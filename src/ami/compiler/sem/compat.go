package sem

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

// typesCompatible returns true when two textual types are considered compatible
// under conservative rules: exact match, wildcard any, container unification
// with any, and generic Event/Error with a single-letter type variable.
func typesCompatible(expected, actual string) bool {
    if expected == "" || expected == "any" || actual == "" || actual == "any" { return true }
    if expected == actual { return true }
    // Optional unwrap: consider Optional<X> compatible with Optional<Y> when X~Y
    if strings.HasPrefix(expected, "Optional<") && strings.HasPrefix(actual, "Optional<") {
        return typesCompatible(innerGeneric(expected), innerGeneric(actual))
    }
    // Union compatibility: actual must be included in expected
    if strings.HasPrefix(expected, "Union<") {
        // parse using types to handle nested generics safely
        et, err1 := types.Parse(expected)
        at, err2 := types.Parse(actual)
        if err1 == nil && err2 == nil {
            return unionContains(et, at)
        }
    }
    // Event/Error/Owned generic compatibility
    if strings.HasPrefix(expected, "Event<") && strings.HasPrefix(actual, "Event<") {
        pe := innerGeneric(expected)
        pa := innerGeneric(actual)
        if isTypeVar(pe) || isTypeVar(pa) { return true }
        return pe == pa
    }
    if strings.HasPrefix(expected, "Error<") && strings.HasPrefix(actual, "Error<") {
        pe := innerGeneric(expected)
        pa := innerGeneric(actual)
        if isTypeVar(pe) || isTypeVar(pa) { return true }
        return pe == pa
    }
    if strings.HasPrefix(expected, "Owned<") && strings.HasPrefix(actual, "Owned<") {
        pe := innerGeneric(expected)
        pa := innerGeneric(actual)
        if isTypeVar(pe) || isTypeVar(pa) { return true }
        return pe == pa
    }
    // slice/set unification
    if strings.HasPrefix(expected, "slice<") && strings.HasPrefix(actual, "slice<") {
        return elemCompatible(expected, actual)
    }
    if strings.HasPrefix(expected, "set<") && strings.HasPrefix(actual, "set<") {
        return elemCompatible(expected, actual)
    }
    // map unification
    if strings.HasPrefix(expected, "map<") && strings.HasPrefix(actual, "map<") {
        ek, ev := keyVal(expected)
        ak, av := keyVal(actual)
        return typesCompatible(ek, ak) && typesCompatible(ev, av)
    }
    return false
}
