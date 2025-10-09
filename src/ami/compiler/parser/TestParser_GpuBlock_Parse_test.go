package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    ast2 "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Verify gpu(...) { ... } blocks parse inside function bodies with attrs and raw source captured.
func TestParser_GpuBlock_Parse(t *testing.T) {
    src := "package app\nfunc F(){ gpu(family=\"metal\", name=\"k\", n=4){ kernel void k(){} } }\n"
    f := &source.File{Name: "u.ami", Content: src}
    af, errs := New(f).ParseFileCollect()
    if af == nil || len(errs) > 0 {
        t.Fatalf("parse errors: %+v", errs)
    }
    // Find the GPU block in the function body
    found := false
    for _, d := range af.Decls {
        fn, ok := d.(*ast2.FuncDecl)
        if !ok || fn.Name != "F" || fn.Body == nil { continue }
        for _, st := range fn.Body.Stmts {
            if g, ok := st.(*ast2.GPUBlockStmt); ok {
                // attrs captured
                if len(g.Attrs) == 0 { t.Fatalf("expected attrs in gpu(...) block") }
                // raw source captured between braces
                if g.Source == "" || !contains(g.Source, "kernel") {
                    t.Fatalf("expected kernel source captured; got: %q", g.Source)
                }
                found = true
            }
        }
    }
    if !found { t.Fatalf("gpu block not found in AST") }
}

// contains is a tiny helper to avoid importing strings in this test.
func contains(s, sub string) bool {
    if len(sub) == 0 { return true }
    for i := 0; i+len(sub) <= len(s); i++ { if s[i:i+len(sub)] == sub { return true } }
    return false
}
