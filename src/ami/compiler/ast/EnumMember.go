package ast

// EnumMember is a single member in an enum with an optional literal value.
// When Value is non-empty, it preserves the literal as written (including quotes for strings)
// so that later phases can distinguish string vs numeric values.
type EnumMember struct {
	Name  string
	Value string
}
