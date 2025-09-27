package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// SetLit represents a literal like set<T>{e1, e2, ...}
type SetLit struct {
    Pos       source.Position
    TypeName  string
    LBrace    source.Position
    Elems     []Expr
    RBrace    source.Position
}

func (*SetLit) isNode() {}
func (*SetLit) isExpr() {}

