package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Decorators(t *testing.T) {
    src := "package app\n@dec(1, \"s\")\nfunc F(){}\n"
    f := (&source.FileSet{}).AddFile("deco.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }
}

