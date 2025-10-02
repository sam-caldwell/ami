package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

type NumberLit struct {
    Pos  source.Position
    Text string
}

func (*NumberLit) isNode() {}
func (*NumberLit) isExpr() {}

