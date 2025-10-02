package ast

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// BinaryExpr represents a binary operation with left/operator/right.
type BinaryExpr struct {
    Pos  source.Position
    Op   token.Kind
    X, Y Expr
}

func (*BinaryExpr) isNode() {}
func (*BinaryExpr) isExpr() {}

