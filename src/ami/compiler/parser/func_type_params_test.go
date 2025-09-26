package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "testing"
)

func findFuncDecl(t *testing.T, decls []astpkg.Node, name string) (astpkg.FuncDecl, bool) {
    t.Helper()
    for _, d := range decls {
        if fd, ok := d.(astpkg.FuncDecl); ok && fd.Name == name {
            return fd, true
        }
    }
    return astpkg.FuncDecl{}, false
}

func TestParser_Func_TypeParams_Single(t *testing.T) {
    src := "package p\nfunc f<T>(x T) { return }\n"
    p := New(src)
    f := p.ParseFile()
    fd, ok := findFuncDecl(t, f.Decls, "f")
    if !ok {
        t.Fatalf("func f not found: %+v", f.Decls)
    }
    if len(fd.TypeParams) != 1 || fd.TypeParams[0].Name != "T" || fd.TypeParams[0].Constraint != "" {
        t.Fatalf("unexpected type params: %+v", fd.TypeParams)
    }
    if len(p.Errors()) != 0 {
        t.Fatalf("unexpected parse errors: %+v", p.Errors())
    }
}

func TestParser_Func_TypeParams_MultipleWithAnyConstraint(t *testing.T) {
    src := "package p\nfunc g<T any, U any>(a T, b U) {}\n"
    p := New(src)
    f := p.ParseFile()
    fd, ok := findFuncDecl(t, f.Decls, "g")
    if !ok {
        t.Fatalf("func g not found: %+v", f.Decls)
    }
    if len(fd.TypeParams) != 2 {
        t.Fatalf("want 2 type params, got: %+v", fd.TypeParams)
    }
    if fd.TypeParams[0].Name != "T" || fd.TypeParams[0].Constraint != "any" {
        t.Fatalf("unexpected first type param: %+v", fd.TypeParams[0])
    }
    if fd.TypeParams[1].Name != "U" || fd.TypeParams[1].Constraint != "any" {
        t.Fatalf("unexpected second type param: %+v", fd.TypeParams[1])
    }
    if len(p.Errors()) != 0 {
        t.Fatalf("unexpected parse errors: %+v", p.Errors())
    }
}

func TestParser_Func_TypeParams_TolerantBadTokens(t *testing.T) {
    // numeric token inside type params is skipped tolerantly
    src := "package p\nfunc h<T, 123, U>(a T) {}\n"
    p := New(src)
    f := p.ParseFile()
    fd, ok := findFuncDecl(t, f.Decls, "h")
    if !ok {
        t.Fatalf("func h not found: %+v", f.Decls)
    }
    if len(fd.TypeParams) != 2 || fd.TypeParams[0].Name != "T" || fd.TypeParams[1].Name != "U" {
        t.Fatalf("tolerant parse failed, got: %+v", fd.TypeParams)
    }
}

