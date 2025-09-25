package types

// OwnedType represents Owned<T> semantics in AMI's type system.
type OwnedType struct{ Elem Type }

func (o OwnedType) String() string { return "Owned<" + o.Elem.String() + ">" }

