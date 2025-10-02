package types

// Optional wraps a single inner type representing an optional value.
type Optional struct{ Inner Type }

func (o Optional) String() string { return "Optional<" + o.Inner.String() + ">" }

