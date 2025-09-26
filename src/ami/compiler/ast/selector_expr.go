package ast

// SelectorExpr represents a qualified selector: receiver.method.
type SelectorExpr struct {
    X   Expr
    Sel string
    Pos Position
}

func (SelectorExpr) isExpr() {}

