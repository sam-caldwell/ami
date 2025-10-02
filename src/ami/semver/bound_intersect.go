package semver

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

