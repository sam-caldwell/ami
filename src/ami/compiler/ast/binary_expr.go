package ast

// BinaryExpr represents a binary operation: X Op Y.
type BinaryExpr struct {
    X   Expr
    Op  string
    Y   Expr
    Pos Position
}

func (BinaryExpr) isExpr() {}

