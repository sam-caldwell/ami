package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

type StringLit struct {
    Pos   source.Position
    Value string
}

func (*StringLit) isNode() {}
func (*StringLit) isExpr() {}

