package ast

// MapElem is a key:value element for map literals.
type MapElem struct {
	Key   Expr
	Value Expr
}
