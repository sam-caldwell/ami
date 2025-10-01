package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Var_StructCapture_InFuncBlock(t *testing.T) {
    code := "package app\nfunc F(){ var x Struct{a:slice<Owned<T>>} }\n"
    var fs source.FileSet
    f := fs.AddFile("z.ami", code)
    p := New(f)
    af, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    fn, _ := af.Decls[0].(*ast.FuncDecl)
    if fn == nil || len(fn.Body.Stmts) == 0 { t.Fatalf("no body stmts") }
    vd, ok := fn.Body.Stmts[0].(*ast.VarDecl)
    if !ok { t.Fatalf("stmt0 not VarDecl: %T", fn.Body.Stmts[0]) }
    if vd.Type != "Struct{a:slice<Owned<T>>}" { t.Fatalf("var type: %q", vd.Type) }
}

