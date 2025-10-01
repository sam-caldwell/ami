package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Param_GenericNestedCapture(t *testing.T) {
    code := "package app\nfunc H(p Outer<Inner<T>>, q int){}\n"
    var fs source.FileSet
    f := fs.AddFile("w.ami", code)
    p := New(f)
    af, errs := p.ParseFileCollect()
    if len(errs) > 0 || af == nil { t.Fatalf("parse errors: %+v", errs) }
    fn, _ := af.Decls[0].(*ast.FuncDecl)
    if fn == nil || len(fn.Params) != 2 { t.Fatalf("params: %+v", fn) }
    if fn.Params[0].Type != "Outer<Inner<T>>" { t.Fatalf("param0 type: %q", fn.Params[0].Type) }
}

