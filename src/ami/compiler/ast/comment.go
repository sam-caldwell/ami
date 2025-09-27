package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Comment represents a source comment (line or block) with raw text.
type Comment struct {
    Pos  source.Position
    Text string
}

