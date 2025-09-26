package ast

// MutBlockStmt wraps a block intended to perform mutating operations.
type MutBlockStmt struct {
    Body     BlockStmt
    Pos      Position
    Comments []Comment
}

func (MutBlockStmt) isStmt() {}

