package ast

// ExprStmt wraps an expression as a statement.
type ExprStmt struct {
    X        Expr
    Pos      Position
    Comments []Comment
}

func (ExprStmt) isStmt() {}

