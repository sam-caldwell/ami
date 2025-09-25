package types

// Pointer represents a pointer type *T.
type Pointer struct{ Base Type }

func (p Pointer) String() string { return "*" + p.Base.String() }

// Slice represents a slice type []T.
type Slice struct{ Elem Type }

func (s Slice) String() string { return "[]" + s.Elem.String() }

// Map represents a map type map[K,V].
type Map struct{ Key, Value Type }

func (m Map) String() string { return "map<" + m.Key.String() + "," + m.Value.String() + ">" }

// Set represents a set type set<T>.
type Set struct{ Elem Type }

func (s Set) String() string { return "set<" + s.Elem.String() + ">" }

// SliceTy models a slice<T> generic type different from []T (language supports both).
type SliceTy struct{ Elem Type }

func (s SliceTy) String() string { return "slice<" + s.Elem.String() + ">" }
