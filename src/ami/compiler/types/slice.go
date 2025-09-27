package types

// Slice represents a bracketed slice form (syntactic form "[]T").
type Slice struct{ Elem Type }

func (s Slice) String() string { return "[]" + s.Elem.String() }

