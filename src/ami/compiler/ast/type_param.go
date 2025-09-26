package ast

// TypeParam represents a single type parameter with an optional constraint.
// Constraint is a tolerant, parser-captured identifier (e.g., "any").
type TypeParam struct {
    Name       string
    Constraint string
}

