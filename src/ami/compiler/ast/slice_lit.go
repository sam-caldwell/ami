package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// SliceLit represents a literal like slice<T>{e1, e2, ...}
type SliceLit struct {
    Pos       source.Position // position of 'slice'
    TypeName  string          // T
    LBrace    source.Position
    Elems     []Expr
    RBrace    source.Position
}

func (*SliceLit) isNode() {}
func (*SliceLit) isExpr() {}

