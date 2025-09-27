package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Decorators must appear immediately before a function declaration, else parser emits an error.
func TestParse_Decorators_Misuse(t *testing.T) {
    src := "package app\n@dec\npipeline P(){}\n"
    f := (&source.FileSet{}).AddFile("bad.ami", src)
    p := New(f)
    _, errs := p.ParseFileCollect()
    if len(errs) == 0 { t.Fatalf("expected decorator misuse error") }
}

