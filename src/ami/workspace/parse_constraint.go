package workspace

import (
    "fmt"
    "regexp"
    "strings"
)

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

