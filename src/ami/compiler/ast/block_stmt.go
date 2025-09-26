package ast

// BlockStmt is a sequence of statements within braces.
type BlockStmt struct {
    Stmts    []Stmt
    Pos      Position
    Comments []Comment
}

func (BlockStmt) isStmt() {}

