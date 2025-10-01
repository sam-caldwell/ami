package scanner

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// New creates a new Scanner for the provided file.
func New(f *source.File) *Scanner { return &Scanner{file: f, offset: 0} }
