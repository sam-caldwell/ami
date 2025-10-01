package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Result_GenericCapture_Multiple(t *testing.T) {
    code := "package app\nfunc H() (Pair<T,U>, string) { return }\n"
    var fs source.FileSet
    f := fs.AddFile("v.ami", code)
    p := New(f)
    af, errs := p.ParseFileCollect()
    if len(errs) > 0 || af == nil { t.Fatalf("parse errors: %+v", errs) }
    fn, _ := af.Decls[0].(*ast.FuncDecl)
    if fn == nil || len(fn.Results) != 2 { t.Fatalf("unexpected results: %+v", fn) }
    if fn.Results[0].Type != "Pair<T,U>" { t.Fatalf("result0: %q", fn.Results[0].Type) }
    if fn.Results[1].Type != "string" { t.Fatalf("result1: %q", fn.Results[1].Type) }
}

