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

func elemCompatible(a, b string) bool {
    ea := innerGeneric(a)
    eb := innerGeneric(b)
    if isTypeVar(ea) || isTypeVar(eb) { return true }
    return typesCompatible(ea, eb)
}

func innerGeneric(s string) string {
    // return content between first '<' and last '>'
    i := strings.IndexByte(s, '<')
    j := strings.LastIndexByte(s, '>')
    if i < 0 || j <= i { return s }
    return s[i+1 : j]
}

func keyVal(s string) (string, string) {
    in := innerGeneric(s)
    // split on first comma
    c := strings.IndexByte(in, ',')
    if c < 0 { return in, "" }
    return strings.TrimSpace(in[:c]), strings.TrimSpace(in[c+1:])
}

func isTypeVar(s string) bool {
    if s == "any" || s == "" { return true }
    // consider single-letter ASCII uppercase as a type variable (T/U/E/etc.)
    return len(s) == 1 && s[0] >= 'A' && s[0] <= 'Z'
}

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
