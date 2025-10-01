package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_EnumDecl_MissingRBrace(t *testing.T) {
    code := "package app\nenum E { A\n"
    f := &source.File{Name: "e3.ami", Content: code}
    p := New(f)
    _, _ = p.ParseFileCollect() // tolerate error; just exercise branch
}

