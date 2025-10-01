package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// ConditionalExpr represents the ternary conditional: Cond ? Then : Else
type ConditionalExpr struct {
    Pos  source.Position
    Cond Expr
    Then Expr
    Else Expr
}

func (*ConditionalExpr) isNode() {}
func (*ConditionalExpr) isExpr() {}

