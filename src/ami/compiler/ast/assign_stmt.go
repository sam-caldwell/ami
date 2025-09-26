package ast

// AssignStmt assigns the result of RHS to LHS.
type AssignStmt struct {
    LHS      Expr
    RHS      Expr
    Pos      Position
    Comments []Comment
}

func (AssignStmt) isStmt() {}

