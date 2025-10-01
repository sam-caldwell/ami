package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Import_BlockForm_WithAndWithoutConstraints(t *testing.T) {
    src := "package app\nimport (\n alpha >= v1.2.3\n \"beta\"\n)\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 2 { t.Fatalf("want 2 imports, got %d", len(file.Decls)) }
    imp0, ok := file.Decls[0].(*ast.ImportDecl)
    if !ok { t.Fatalf("decl0 not ImportDecl: %T", file.Decls[0]) }
    if imp0.Path != "alpha" || imp0.Constraint != ">= v1.2.3" { t.Fatalf("decl0 unexpected: %+v", imp0) }
    imp1, ok := file.Decls[1].(*ast.ImportDecl)
    if !ok { t.Fatalf("decl1 not ImportDecl: %T", file.Decls[1]) }
    if imp1.Path != "beta" || imp1.Constraint != "" { t.Fatalf("decl1 unexpected: %+v", imp1) }
}

