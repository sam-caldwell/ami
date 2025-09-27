package semver

import (
    "fmt"
    "regexp"
    "strings"
)

var (
    reVersion    = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)
    reConstraint = regexp.MustCompile(`^(\^|~|>=|>)?\s*(v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)$|^==latest$`)
)

// Constraint represents a parsed version constraint.
type Constraint struct {
    Op      string // one of ^, ~, >=, >, or "" for exact
    Version string // normalized without surrounding spaces; may be empty when Latest is true
    Latest  bool   // true when constraint is ==latest
}

// ParseConstraint parses a version constraint string.
func ParseConstraint(s string) (Constraint, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return Constraint{}, fmt.Errorf("empty constraint")
    }
    if s == "==latest" {
        return Constraint{Latest: true}, nil
    }
    m := reConstraint.FindStringSubmatch(s)
    if m == nil {
        return Constraint{}, fmt.Errorf("invalid constraint: %s", s)
    }
    op := strings.TrimSpace(m[1])
    ver := strings.TrimSpace(m[2])
    if !reVersion.MatchString(ver) {
        return Constraint{}, fmt.Errorf("invalid version: %s", ver)
    }
    // normalize to leading v
    if !strings.HasPrefix(ver, "v") { ver = "v" + ver }
    return Constraint{Op: op, Version: ver}, nil
}

// ValidateConstraint returns true when the constraint string is valid per ParseConstraint.
func ValidateConstraint(s string) bool {
    _, err := ParseConstraint(s)
    return err == nil
}

// ValidateVersion returns true when s is a valid semver (optionally prefixed with v).
func ValidateVersion(s string) bool { return reVersion.MatchString(s) }

