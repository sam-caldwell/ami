package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// TypeParam represents a function type parameter (scaffold).
// Constraint is an optional identifier such as "any".
type TypeParam struct {
    Pos        source.Position
    Name       string
    NamePos    source.Position
    Constraint string
}

func (*TypeParam) isNode() {}

