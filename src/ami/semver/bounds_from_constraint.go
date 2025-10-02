package semver

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

