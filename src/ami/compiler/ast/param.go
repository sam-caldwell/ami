package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Param represents a single function parameter.
type Param struct {
    Pos  source.Position
    Name string
    Type string
    // TypePos is the position of the type token (start of the type name).
    // This enables more precise diagnostics pointing at the type, not just the name.
    TypePos source.Position
    Leading []Comment
}
