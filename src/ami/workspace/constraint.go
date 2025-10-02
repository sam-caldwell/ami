package workspace

// Constraint captures an operator and a target version.
// For OpLatest, Version is empty.
type Constraint struct {
    Op      Op
    Version string
    v       semver
}

