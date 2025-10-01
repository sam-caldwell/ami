package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_Decorators_Single_NoArgs(t *testing.T) {
    src := "package p\n@trace\nfunc F() {}\n"
    var fs source.FileSet
    f := fs.AddFile("f.ami", src)
    p := New(f)
    file, err := p.ParseFile()
    if err != nil {
        t.Fatalf("parse error: %v", err)
    }
    if len(file.Decls) != 1 {
        t.Fatalf("expected 1 decl, got %d", len(file.Decls))
    }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok {
        t.Fatalf("decl 0 not a func")
    }
    if len(fn.Decorators) != 1 {
        t.Fatalf("expected 1 decorator, got %d", len(fn.Decorators))
    }
    if fn.Decorators[0].Name != "trace" {
        t.Fatalf("decorator name: %s", fn.Decorators[0].Name)
    }
}

