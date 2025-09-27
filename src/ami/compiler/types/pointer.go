package types

// Pointer represents a pointer to an element type (syntactic form "*T").
type Pointer struct{ Elem Type }

func (p Pointer) String() string { return "*" + p.Elem.String() }

