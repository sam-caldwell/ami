package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Param represents a single function parameter.
type Param struct {
    Pos  source.Position
    Name string
    Type string
}
