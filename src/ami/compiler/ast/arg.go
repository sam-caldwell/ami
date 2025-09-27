package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Arg represents a simple argument in a call (identifier or string literal).
type Arg struct {
    Pos      source.Position
    Text     string
    IsString bool
}

