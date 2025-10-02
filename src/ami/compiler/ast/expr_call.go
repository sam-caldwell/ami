package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

type CallExpr struct {
    Pos     source.Position
    Name    string
    NamePos source.Position
    LParen  source.Position
    Args    []Expr
    RParen  source.Position
}

func (*CallExpr) isNode() {}
func (*CallExpr) isExpr() {}

