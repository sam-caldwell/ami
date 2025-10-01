package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_EnumDecl_Blanks_And_Recovery(t *testing.T) {
    code := "package app\nenum E { , A, , B, }\n"
    f := &source.File{Name: "e2.ami", Content: code}
    p := New(f)
    af, _ := p.ParseFileCollect()
    var ed *ast.EnumDecl
    for _, d := range af.Decls { if v, ok := d.(*ast.EnumDecl); ok { ed = v; break } }
    if ed == nil || len(ed.Members) < 4 { t.Fatalf("enum members not parsed: %+v", ed) }
}

