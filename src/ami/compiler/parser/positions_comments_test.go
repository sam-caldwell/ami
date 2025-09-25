package parser

import (
    "strings"
    "testing"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// helper to find first node of a given type
func findDecl[T any](t *testing.T, decls []astpkg.Node) (T, bool) {
    t.Helper()
    var zero T
    for _, d := range decls {
        if v, ok := d.(T); ok { return v, true }
    }
    return zero, false
}

func TestParser_PositionsAndComments_Attached(t *testing.T) {
    src := `// before pragma
#pragma test:case Demo
// before import
import "x/y"
// before enum
enum E { A = 1 }
// before struct
struct S { F string }
// before func
func f(ctx Context, ev Event<string>, st State) Event<string> { }
// before pipeline
pipeline P {
  // step ingress
  Ingress(cfg).Transform(x).Egress(cfg)
}
`
    p := New(src)
    f := p.ParseFile()

    if len(f.Directives) == 0 { t.Fatalf("expected at least one directive") }
    if len(f.Directives[0].Comments) == 0 { t.Fatalf("expected comments on directive") }
    if !strings.Contains(f.Directives[0].Comments[0].Text, "before pragma") { t.Fatalf("directive comment missing text: %+v", f.Directives[0].Comments) }
    if f.Directives[0].Pos.Line <= 0 { t.Fatalf("expected directive position to be set") }

    // ImportDecl
    var imp astpkg.ImportDecl
    ok := false
    for _, d := range f.Decls {
        if v, is := d.(astpkg.ImportDecl); is { imp = v; ok = true; break }
    }
    if !ok { t.Fatalf("expected import decl present") }
    if len(imp.Comments) == 0 || !strings.Contains(imp.Comments[0].Text, "before import") { t.Fatalf("expected comment before import; got %+v", imp.Comments) }
    if imp.Pos.Line <= 0 { t.Fatalf("expected import position set") }

    // EnumDecl
    var en astpkg.EnumDecl
    for _, d := range f.Decls { if v, is := d.(astpkg.EnumDecl); is { en = v } }
    if en.Name == "" { t.Fatalf("expected enum decl present") }
    if len(en.Comments) == 0 || !strings.Contains(en.Comments[0].Text, "before enum") { t.Fatalf("expected enum comment; got %+v", en.Comments) }
    if en.Pos.Line <= 0 { t.Fatalf("expected enum pos set") }

    // StructDecl
    var st astpkg.StructDecl
    for _, d := range f.Decls { if v, is := d.(astpkg.StructDecl); is { st = v } }
    if st.Name == "" { t.Fatalf("expected struct decl present") }
    if len(st.Comments) == 0 || !strings.Contains(st.Comments[0].Text, "before struct") { t.Fatalf("expected struct comment; got %+v", st.Comments) }
    if st.Pos.Line <= 0 { t.Fatalf("expected struct pos set") }

    // FuncDecl
    var fn astpkg.FuncDecl
    for _, d := range f.Decls { if v, is := d.(astpkg.FuncDecl); is { fn = v } }
    if fn.Name == "" { t.Fatalf("expected func decl present") }
    if len(fn.Comments) == 0 || !strings.Contains(fn.Comments[0].Text, "before func") { t.Fatalf("expected func comment; got %+v", fn.Comments) }
    if fn.Pos.Line <= 0 { t.Fatalf("expected func pos set") }

    // PipelineDecl and NodeCall comment
    var pd astpkg.PipelineDecl
    for _, d := range f.Decls { if v, is := d.(astpkg.PipelineDecl); is { pd = v } }
    if pd.Name == "" { t.Fatalf("expected pipeline decl present") }
    if len(pd.Steps) == 0 { t.Fatalf("expected pipeline steps") }
    if len(pd.Steps[0].Comments) == 0 || !strings.Contains(pd.Steps[0].Comments[0].Text, "step ingress") { t.Fatalf("expected step comment on first node call; got %+v", pd.Steps[0].Comments) }
    if pd.Pos.Line <= 0 { t.Fatalf("expected pipeline pos set") }
}
