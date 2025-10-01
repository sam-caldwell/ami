package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Func_Assign_Binary_Defer(t *testing.T) {
    src := "package app\nfunc G(){ x = 1+2*3; defer Alpha(); }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("parse: %v", err) }
}

