package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Attr represents a simple attribute: Name or Name(args...).
type Attr struct {
    Pos  source.Position
    Name string
    Args []Arg
}

