package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Import_WithVersionConstraint(t *testing.T) {
    src := "package app\nimport alpha >= v1.2.3\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    imp, ok := file.Decls[0].(*ast.ImportDecl)
    if !ok { t.Fatalf("first decl not ImportDecl: %T", file.Decls[0]) }
    if imp.Path != "alpha" { t.Fatalf("path: %q", imp.Path) }
    if imp.Constraint != ">= v1.2.3" { t.Fatalf("constraint: %q", imp.Constraint) }
}

