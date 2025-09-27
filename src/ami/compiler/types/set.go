package types

// Set represents a set generic form (semantic form "set<T>").
type Set struct{ Elem Type }

func (s Set) String() string { return "set<" + s.Elem.String() + ">" }

