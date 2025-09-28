package ast

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// UnaryExpr represents a unary operator expression like `!x` or `-x`.
type UnaryExpr struct {
    Pos source.Position
    Op  token.Kind
    X   Expr
}

func (*UnaryExpr) isNode() {}
func (*UnaryExpr) isExpr() {}

