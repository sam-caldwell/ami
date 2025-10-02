package sem

import cty "github.com/sam-caldwell/ami/src/ami/compiler/types"

// payloadWithin reports whether actual equals expected or, when expected is a Union, whether
// actual is one of the union alternatives.
func payloadWithin(expected, actual cty.Type) bool {
    // Exact structural match
    if cty.Equal(expected, actual) { return true }

    // Optional covariance: Optional<X> accepts X or Optional<Y> when Y ∈ X (recursively).
    if eo, ok := expected.(cty.Optional); ok {
        // If actual is Optional<Y>, compare Y to X; else compare actual to X directly.
        if ao, ok := actual.(cty.Optional); ok { return payloadWithin(eo.Inner, ao.Inner) }
        return payloadWithin(eo.Inner, actual)
    }

    // Union membership: expected is Union, actual must be a member.
    if eu, ok := expected.(cty.Union); ok {
        for _, alt := range eu.Alts { if payloadWithin(alt, actual) { return true } }
        return false
    }

    // Struct width subtyping: expected struct fields must be present and compatible in actual.
    if es, ok := expected.(cty.Struct); ok {
        if as, ok := actual.(cty.Struct); ok {
            for k, et := range es.Fields {
                at, ok := as.Fields[k]
                if !ok { return false }
                // Allow Optional wrappers on actual fields when expected is non-Optional, by unwrapping once here.
                if ao, ok := at.(cty.Optional); ok { at = ao.Inner }
                if !payloadWithin(et, at) { return false }
            }
            return true
        }
        return false
    }

    // Container variance: same generic name with compatible type arguments.
    // Supported forms via Parse: slice<T>, set<T>, map<K,V>, Owned<T>, Error<E>, etc.
    if eg, ok := expected.(cty.Generic); ok {
        if ag, ok := actual.(cty.Generic); ok && eg.Name == ag.Name && len(eg.Args) == len(ag.Args) {
            for i := range eg.Args { if !payloadWithin(eg.Args[i], ag.Args[i]) { return false } }
            return true
        }
    }

    // No match by structure.
    return false
}

