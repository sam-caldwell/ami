package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_ImportConstraint_QuotedMissingV(t *testing.T) {
    src := "package app\nimport \"x\" >= \"1.2.3\"\n"
    f := (&source.FileSet{}).AddFile("icnv.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

