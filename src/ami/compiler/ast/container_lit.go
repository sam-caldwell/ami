package ast

// ContainerLit represents slice, set, or map literal expressions.
type ContainerLit struct {
    Kind     string
    TypeArgs []TypeRef
    Elems    []Expr
    MapElems []MapElem
    Pos      Position
}

func (ContainerLit) isExpr() {}

// MapElem is a key:value element for map literals.
type MapElem struct {
    Key   Expr
    Value Expr
}

