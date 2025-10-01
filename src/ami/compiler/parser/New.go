package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// New creates a new Parser bound to the provided file.
func New(f *source.File) *Parser {
    p := &Parser{s: scanner.New(f)}
    p.advance()
    return p
}

