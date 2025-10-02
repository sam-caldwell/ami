package workspace

// Op denotes a version constraint operator.
type Op int

// Operators for version constraints (typed to Op).
const (
    OpExact Op = iota
    OpCaret
    OpTilde
    OpGT
    OpGTE
    OpLatest
)

