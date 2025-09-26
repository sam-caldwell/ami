package ast

// Expr is the common interface implemented by all expressions.
type Expr interface{ isExpr() }
