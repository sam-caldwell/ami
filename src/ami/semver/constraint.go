package semver

// Constraint represents a parsed version constraint.
type Constraint struct {
    Op      string // one of ^, ~, >=, >, or "" for exact
    Version string // normalized without surrounding spaces; may be empty when Latest is true
    Latest  bool   // true when constraint is ==latest
}
 
