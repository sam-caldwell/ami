package sem

import (
    "strings"
    cty "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

// compatibleEventTypes returns true when downstream 'expected' event type is compatible with
// upstream 'actual'. It first applies simple text compatibility, then augments for Event<Union<...>>
// where membership of the actual payload type within the expected union is considered compatible.
func compatibleEventTypes(expected, actual string) bool {
    // fast path using existing textual rules
    if typesCompatible(expected, actual) { return true }
    // structural check for Event<...>
    if !strings.HasPrefix(expected, "Event<") || !strings.HasPrefix(actual, "Event<") { return false }
    et, err1 := cty.Parse(expected)
    at, err2 := cty.Parse(actual)
    if err1 != nil || err2 != nil { return false }
    // Unwrap Event payloads
    eg, ok1 := et.(cty.Generic)
    ag, ok2 := at.(cty.Generic)
    if !ok1 || !ok2 || eg.Name != "Event" || ag.Name != "Event" || len(eg.Args) != 1 || len(ag.Args) != 1 { return false }
    exp := eg.Args[0]
    act := ag.Args[0]
    // Direct structural equality
    if cty.Equal(exp, act) { return true }
    // If expected is Optional<X>, allow actual X or Optional<X> (including union membership beneath Optional).
    if eo, ok := exp.(cty.Optional); ok {
        inner := eo.Inner
        // If actual is Optional<Y>, compare Y against inner; else compare actual directly against inner.
        if ao, ok := act.(cty.Optional); ok { return payloadWithin(inner, ao.Inner) }
        return payloadWithin(inner, act)
    }
    // Otherwise check if expected is a Union and actual is a member.
    return payloadWithin(exp, act)
}

