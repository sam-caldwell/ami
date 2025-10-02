package workspace

// Requirement represents a remote dependency declared by a workspace import.
// Local (./) imports are excluded from this representation.
type Requirement struct {
    Name       string
    Constraint Constraint
}

