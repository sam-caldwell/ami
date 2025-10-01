package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_ContainerLiterals_Slice_Set_Map(t *testing.T) {
    src := "package app\nfunc F(){ x = slice<T>{1,2}; y = set<T>{\"a\"}; z = map<K,V>{\"k\": 3} }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl type: %T", file.Decls[0]) }
    if fn.Body == nil || len(fn.Body.Stmts) < 3 { t.Fatalf("expected 3+ statements") }
    // We just ensure parsing didn’t error; deeper assertions would require walking Exprs.
}

