package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Package describes a single package input: name and files.
type Package struct {
    Name  string
    Files *source.FileSet
}

