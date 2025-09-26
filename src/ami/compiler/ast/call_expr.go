package ast

// CallExpr represents a function or method call with argument expressions.
type CallExpr struct {
    Fun  Expr
    Args []Expr
    Pos  Position
}

func (CallExpr) isExpr() {}

