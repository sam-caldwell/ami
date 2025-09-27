package semver

// Bound represents a half-open or closed interval over semantic versions.
// Upper == nil means unbounded above. Inclusivity flags control endpoint inclusion.
type Bound struct {
    Lower          Version
    LowerInclusive bool
    Upper          *Version
    UpperInclusive bool
}

// Bounds returns an equivalent bound interval for a constraint.
// Returns ok=false for unsupported/Latest constraints.
func Bounds(c Constraint) (Bound, bool) {
    if c.Latest { return Bound{}, false }
    v, err := ParseVersion(c.Version)
    if err != nil { return Bound{}, false }
    switch c.Op {
    case "": // exact
        vv := v
        return Bound{Lower: v, LowerInclusive: true, Upper: &vv, UpperInclusive: true}, true
    case ">=":
        return Bound{Lower: v, LowerInclusive: true, Upper: nil, UpperInclusive: false}, true
    case ">":
        return Bound{Lower: v, LowerInclusive: false, Upper: nil, UpperInclusive: false}, true
    case "<=":
        // Lower bound is zero version inclusive; upper is v inclusive
        zero := Version{Major: 0, Minor: 0, Patch: 0}
        vv := v
        return Bound{Lower: zero, LowerInclusive: true, Upper: &vv, UpperInclusive: true}, true
    case "<":
        // Lower bound is zero version inclusive; upper is v exclusive
        zero := Version{Major: 0, Minor: 0, Patch: 0}
        vv := v
        return Bound{Lower: zero, LowerInclusive: true, Upper: &vv, UpperInclusive: false}, true
    case "^":
        up := Version{Major: v.Major + 1}
        return Bound{Lower: v, LowerInclusive: true, Upper: &up, UpperInclusive: false}, true
    case "~":
        up := Version{Major: v.Major, Minor: v.Minor + 1}
        return Bound{Lower: v, LowerInclusive: true, Upper: &up, UpperInclusive: false}, true
    default:
        return Bound{}, false
    }
}

// Intersect returns the intersection of two bounds. ok=false when empty.
func Intersect(a, b Bound) (Bound, bool) {
    // Lower: pick the greater
    lower := a.Lower
    lowerInc := a.LowerInclusive
    if Compare(b.Lower, a.Lower) > 0 {
        lower = b.Lower
        lowerInc = b.LowerInclusive
    } else if Compare(b.Lower, a.Lower) == 0 {
        // both equal, inclusive only if both inclusive
        lowerInc = a.LowerInclusive && b.LowerInclusive
    }
    // Upper: pick the lesser (nil means +inf)
    var upper *Version
    upperInc := false
    if a.Upper == nil {
        upper = b.Upper
        upperInc = b.UpperInclusive
    } else if b.Upper == nil {
        upper = a.Upper
        upperInc = a.UpperInclusive
    } else {
        if Compare(*a.Upper, *b.Upper) < 0 {
            upper = a.Upper
            upperInc = a.UpperInclusive
        } else if Compare(*a.Upper, *b.Upper) > 0 {
            upper = b.Upper
            upperInc = b.UpperInclusive
        } else {
            // equal uppers: inclusive only if both inclusive
            upper = a.Upper
            upperInc = a.UpperInclusive && b.UpperInclusive
        }
    }
    // Empty check
    if upper != nil {
        cmp := Compare(lower, *upper)
        if cmp > 0 { return Bound{}, false }
        if cmp == 0 && !(lowerInc && upperInc) { return Bound{}, false }
    }
    return Bound{Lower: lower, LowerInclusive: lowerInc, Upper: upper, UpperInclusive: upperInc}, true
}

// Contains reports whether v lies within bound b.
func Contains(b Bound, v Version) bool {
    // lower check
    cmp := Compare(v, b.Lower)
    if cmp < 0 || (cmp == 0 && !b.LowerInclusive) { return false }
    if b.Upper == nil { return true }
    cmp = Compare(v, *b.Upper)
    if cmp > 0 || (cmp == 0 && !b.UpperInclusive) { return false }
    return true
}
