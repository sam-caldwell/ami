package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

type IdentExpr struct {
    Pos  source.Position
    Name string
}

func (*IdentExpr) isNode() {}
func (*IdentExpr) isExpr() {}

