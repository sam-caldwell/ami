package ast

// Expr is implemented by all expression nodes.
type Expr interface {
	isExpr()
}
