package semver

// Satisfies reports whether version string v satisfies constraint c.
// Supports operators: ^, ~, >=, >; exact ("" op) is equality; ==latest always returns true.
func Satisfies(v string, c Constraint) bool {
    if c.Latest {
        return true
    }
    // Normalize version (allow leading v)
    if !ValidateVersion(v) { return false }
    pv, err := ParseVersion(v)
    if err != nil { return false }
    // Normalize constraint version to include leading v (ParseConstraint does this)
    cv, err := ParseVersion(c.Version)
    if err != nil { return false }
    switch c.Op {
    case "": // exact
        return Compare(pv, cv) == 0
    case ">=":
        return Compare(pv, cv) >= 0
    case ">":
        return Compare(pv, cv) > 0
    case "^":
        // Compatible with same major: >= cv and < (cv.Major+1).0.0
        upper := Version{Major: cv.Major + 1}
        return Compare(pv, cv) >= 0 && Compare(pv, upper) < 0
    case "~":
        // Approximately equivalent: >= cv and < cv.Major.(cv.Minor+1).0
        upper := Version{Major: cv.Major, Minor: cv.Minor + 1}
        return Compare(pv, cv) >= 0 && Compare(pv, upper) < 0
    default:
        return false
    }
}

