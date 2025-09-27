package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_ImportConstraint_Forms(t *testing.T) {
    src := "package app\nimport foo >= v1.2.3\nimport (\n \"bar\" >= v0.1.0\n)\n"
    f := (&source.FileSet{}).AddFile("ic.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }
}

