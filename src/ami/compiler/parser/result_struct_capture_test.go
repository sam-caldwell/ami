package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Result_StructCapture(t *testing.T) {
    code := "package app\nfunc H() (Struct{a:slice<Owned<T>>}) { return }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", code)
    p := New(f)
    af, errs := p.ParseFileCollect()
    if len(errs) > 0 || af == nil {
        t.Fatalf("parse errors: %+v", errs)
    }
    if len(af.Decls) == 0 {
        t.Fatalf("no decls")
    }
    fn, _ := af.Decls[0].(*ast.FuncDecl)
    if fn == nil || len(fn.Results) == 0 {
        t.Fatalf("no results")
    }
    got := fn.Results[0].Type
    want := "Struct{a:slice<Owned<T>>}"
    if got != want {
        t.Fatalf("result type capture mismatch: got %q want %q", got, want)
    }
}

