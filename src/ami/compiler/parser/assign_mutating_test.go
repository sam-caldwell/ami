package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestParser_Assign_MutatingAndNormal(t *testing.T) {
    src := "package app\nfunc F(){ x = 1; *y = 2 }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok || fn.Body == nil { t.Fatalf("no func/body") }
    if len(fn.Body.Stmts) != 2 { t.Fatalf("want 2 stmts, got %d", len(fn.Body.Stmts)) }
    // normal assign
    a0, ok := fn.Body.Stmts[0].(*ast.AssignStmt)
    if !ok || a0.Mutating { t.Fatalf("stmt0 not normal assign: %#v", fn.Body.Stmts[0]) }
    if a0.Name != "x" { t.Fatalf("a0 name: %s", a0.Name) }
    // mutating assign
    a1, ok := fn.Body.Stmts[1].(*ast.AssignStmt)
    if !ok || !a1.Mutating { t.Fatalf("stmt1 not mutating assign: %#v", fn.Body.Stmts[1]) }
    if a1.Name != "y" || a1.StarPos.Line == 0 { t.Fatalf("mut assign fields: %#v", a1) }
}

