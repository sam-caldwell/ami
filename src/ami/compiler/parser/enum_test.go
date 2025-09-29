package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_EnumDecl_Parse(t *testing.T) {
    code := "package app\nenum Color { Red, Green, Blue }\n"
    f := &source.File{Name: "e.ami", Content: code}
    p := New(f)
    af, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    var found bool
    for _, d := range af.Decls {
        if ed, ok := d.(*ast.EnumDecl); ok {
            if ed.Name != "Color" { t.Fatalf("enum name: %s", ed.Name) }
            if len(ed.Members) != 3 || ed.Members[0].Name != "Red" { t.Fatalf("members: %+v", ed.Members) }
            found = true
        }
    }
    if !found { t.Fatalf("enum decl not found") }
}

