package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Func_TypeParams_NoConstraint(t *testing.T) {
    src := "package app\nfunc F<T>(a T) {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl is %T", file.Decls[0]) }
    if len(fn.TypeParams) != 1 || fn.TypeParams[0].Name != "T" || fn.TypeParams[0].Constraint != "" {
        t.Fatalf("unexpected type params: %+v", fn.TypeParams)
    }
}

