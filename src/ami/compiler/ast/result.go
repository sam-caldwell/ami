package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Result represents a single return item type (scaffold).
type Result struct {
    Pos  source.Position
    Type string
}

