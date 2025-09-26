package ast

// ReturnStmt returns from a function with zero or more expressions.
type ReturnStmt struct {
    Results  []Expr
    Pos      Position
    Comments []Comment
}

func (ReturnStmt) isStmt() {}

