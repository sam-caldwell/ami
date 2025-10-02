package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// EnumMember is a single enum member name and its position.
type EnumMember struct {
	Pos  source.Position
	Name string
}
