package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Types_PrimitiveKeywords_AsTypes(t *testing.T) {
    src := "package app\nfunc F(a int) (string) { var y string; return y }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl is %T", file.Decls[0]) }
    if len(fn.Params) != 1 || fn.Params[0].Type != "int" { t.Fatalf("param type: %+v", fn.Params) }
    if len(fn.Results) != 1 || fn.Results[0].Type != "string" { t.Fatalf("results: %+v", fn.Results) }
    // var y string;
    if fn.Body == nil || len(fn.Body.Stmts) == 0 { t.Fatalf("no body stmts") }
    if vd, ok := fn.Body.Stmts[0].(*ast.VarDecl); ok {
        if vd.Type != "string" { t.Fatalf("var type: %q", vd.Type) }
    } else {
        t.Fatalf("stmt0 not VarDecl: %T", fn.Body.Stmts[0])
    }
}

