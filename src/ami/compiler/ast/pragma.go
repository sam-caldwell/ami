package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Pragma captures a single '#pragma ...' directive as raw text with a position.
// Parsing/interpretation is deferred to consumers (e.g., test harness or linter).
type Pragma struct {
    Pos  source.Position
    Text string
}

