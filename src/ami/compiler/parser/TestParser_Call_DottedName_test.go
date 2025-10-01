package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Call_DottedName(t *testing.T) {
    src := "package app\nfunc F(){ a.b() }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok || fn.Body == nil { t.Fatalf("no func decl/body") }
    es, ok := fn.Body.Stmts[0].(*ast.ExprStmt)
    if !ok { t.Fatalf("stmt0 not ExprStmt: %T", fn.Body.Stmts[0]) }
    ce, ok := es.X.(*ast.CallExpr)
    if !ok { t.Fatalf("expr not CallExpr: %T", es.X) }
    if ce.Name != "a.b" { t.Fatalf("call name: %q", ce.Name) }
}

