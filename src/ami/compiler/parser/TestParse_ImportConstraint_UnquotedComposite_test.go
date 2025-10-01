package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_ImportConstraint_UnquotedComposite(t *testing.T) {
    src := "package app\nimport foo >= v1.2.3-rc.1\n"
    f := (&source.FileSet{}).AddFile("icu.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil {
        t.Fatalf("ParseFile: %v", err)
    }
}

