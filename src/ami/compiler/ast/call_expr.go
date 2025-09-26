package ast

// CallExpr represents a function or method call with argument expressions.
type CallExpr struct {
    Fun  Expr
    Args []Expr
    TypeArgs []TypeRef
    Pos  Position
}

func (CallExpr) isExpr() {}
