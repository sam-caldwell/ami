package ast

// ContainerLit represents slice, set, or map literal expressions.
type ContainerLit struct {
	Kind     string
	TypeArgs []TypeRef
	Elems    []Expr
	MapElems []MapElem
	Pos      Position
}
