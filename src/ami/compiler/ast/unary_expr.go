package ast

// UnaryExpr represents a unary operation like "*x".
type UnaryExpr struct {
    Op  string
    X   Expr
    Pos Position
}

func (UnaryExpr) isExpr() {}

