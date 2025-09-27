package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

type MapElem struct {
    Key Expr
    Val Expr
}

// MapLit represents a literal like map<K,V>{k1: v1, ...}
type MapLit struct {
    Pos       source.Position
    KeyType   string
    ValType   string
    LBrace    source.Position
    Elems     []MapElem
    RBrace    source.Position
}

func (*MapLit) isNode() {}
func (*MapLit) isExpr() {}

