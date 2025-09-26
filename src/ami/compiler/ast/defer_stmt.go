package ast

// DeferStmt represents a deferred execution of an expression (usually a call).
type DeferStmt struct {
    X        Expr
    Pos      Position
    Comments []Comment
}

func (DeferStmt) isStmt() {}

