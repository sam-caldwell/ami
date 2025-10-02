package types

// Named represents a named type or type variable (e.g., user type or single-letter T).
type Named struct{ Name string }

func (n Named) String() string { return n.Name }

