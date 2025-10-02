package workspace

import "strings"

// Satisfies returns whether version v matches the constraint.
func Satisfies(version string, c Constraint) bool {
    if c.Op == OpLatest { return true }
    v, err := parseSemver(trimV(strings.TrimSpace(version)))
    if err != nil { return false }
    switch c.Op {
    case OpExact:
        return cmpSemver(v, c.v) == 0
    case OpGT:
        return cmpSemver(v, c.v) > 0
    case OpGTE:
        return cmpSemver(v, c.v) >= 0
    case OpTilde:
        // >= X.Y.Z and < X.(Y+1).0
        upper := semver{Major: c.v.Major, Minor: c.v.Minor + 1}
        return cmpSemver(v, c.v) >= 0 && cmpSemver(v, upper) < 0
    case OpCaret:
        if c.v.Major > 0 {
            // >= X.Y.Z and < (X+1).0.0
            upper := semver{Major: c.v.Major + 1}
            return cmpSemver(v, c.v) >= 0 && cmpSemver(v, upper) < 0
        }
        if c.v.Minor > 0 {
            // Major==0: lock minor
            upper := semver{Major: 0, Minor: c.v.Minor + 1}
            return cmpSemver(v, c.v) >= 0 && cmpSemver(v, upper) < 0
        }
        // Major==0, Minor==0: lock patch
        upper := semver{Major: 0, Minor: 0, Patch: c.v.Patch + 1}
        return cmpSemver(v, c.v) >= 0 && cmpSemver(v, upper) < 0
    default:
        return false
    }
}

