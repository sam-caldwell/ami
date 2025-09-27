package types

// SliceTy represents the generic slice type form (semantic form "slice<T>").
type SliceTy struct{ Elem Type }

func (s SliceTy) String() string { return "slice<" + s.Elem.String() + ">" }

