package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Func_TypeParams_WithConstraint_AndMultiple(t *testing.T) {
    src := "package app\nfunc F<T any, U>(a T, b U) {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl is %T", file.Decls[0]) }
    if len(fn.TypeParams) != 2 { t.Fatalf("want 2 type params, got %d", len(fn.TypeParams)) }
    if fn.TypeParams[0].Name != "T" || fn.TypeParams[0].Constraint != "any" { t.Fatalf("tp0: %+v", fn.TypeParams[0]) }
    if fn.TypeParams[1].Name != "U" || fn.TypeParams[1].Constraint != "" { t.Fatalf("tp1: %+v", fn.TypeParams[1]) }
}

