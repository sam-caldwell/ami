package types

// ObjectKind classifies a declared name.
type ObjectKind int

const (
	ObjVar ObjectKind = iota
	ObjFunc
	ObjType
)

// Object represents a declared name bound to a kind and type.
type Object struct {
	Kind ObjectKind
	Name string
	Type Type
}
