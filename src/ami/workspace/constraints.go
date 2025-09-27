package workspace

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
)

// Op denotes a version constraint operator.
type Op int

const (
    OpExact Op = iota
    OpCaret
    OpTilde
    OpGT
    OpGTE
    OpLatest
)

// Constraint captures an operator and a target version.
// For OpLatest, Version is empty.
type Constraint struct {
    Op      Op
    Version string
    v       semver
}

var reConstraint = regexp.MustCompile(`^\s*(==latest|\^|~|>=|>|v?)(\s*[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?)?\s*$`)

// ParseConstraint parses a constraint string like "^1.2.3", ">= 1.2.3", "v1.2.3", or "==latest".
// Unsupported operators (e.g., "<=") return an error.
func ParseConstraint(s string) (Constraint, error) {
    m := reConstraint.FindStringSubmatch(s)
    if m == nil { return Constraint{}, fmt.Errorf("invalid constraint: %q", s) }
    opRaw := strings.TrimSpace(m[1])
    verRaw := strings.TrimSpace(m[2])
    var c Constraint
    switch opRaw {
    case "^": c.Op = OpCaret
    case "~": c.Op = OpTilde
    case ">": c.Op = OpGT
    case ">=": c.Op = OpGTE
    case "==latest":
        c.Op = OpLatest
        return c, nil
    case "v", "":
        c.Op = OpExact
    default:
        return Constraint{}, fmt.Errorf("unsupported operator: %s", opRaw)
    }
    if verRaw == "" { return Constraint{}, fmt.Errorf("missing version in constraint") }
    c.Version = trimV(verRaw)
    v, err := parseSemver(c.Version)
    if err != nil { return Constraint{}, err }
    c.v = v
    return c, nil
}

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

// semver is a minimal semantic version model.
type semver struct {
    Major int
    Minor int
    Patch int
    Pre   string // optional pre-release; ignored for range bounds
}

func trimV(s string) string { return strings.TrimPrefix(strings.TrimSpace(s), "v") }

func parseSemver(s string) (semver, error) {
    // s = MAJOR.MINOR.PATCH[-PRERELEASE]
    parts := strings.SplitN(s, "-", 2)
    nums := strings.Split(parts[0], ".")
    if len(nums) != 3 { return semver{}, fmt.Errorf("invalid semver: %q", s) }
    maj, err := strconv.Atoi(nums[0]); if err != nil { return semver{}, err }
    min, err := strconv.Atoi(nums[1]); if err != nil { return semver{}, err }
    pat, err := strconv.Atoi(nums[2]); if err != nil { return semver{}, err }
    v := semver{Major: maj, Minor: min, Patch: pat}
    if len(parts) == 2 { v.Pre = parts[1] }
    return v, nil
}

func cmpSemver(a, b semver) int {
    if a.Major != b.Major { if a.Major < b.Major { return -1 } ; return 1 }
    if a.Minor != b.Minor { if a.Minor < b.Minor { return -1 } ; return 1 }
    if a.Patch != b.Patch { if a.Patch < b.Patch { return -1 } ; return 1 }
    // Pre-release: treat empty (release) > any pre-release
    if a.Pre == b.Pre { return 0 }
    if a.Pre == "" { return 1 }
    if b.Pre == "" { return -1 }
    // Fallback lexical compare for prerelease
    if a.Pre < b.Pre { return -1 }
    if a.Pre > b.Pre { return 1 }
    return 0
}

