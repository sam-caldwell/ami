package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    ast2 "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Verify multiple gpu(...) blocks and different families parse in order.
func TestParser_GpuBlock_Multiple_And_OpenCL(t *testing.T) {
    src := "package app\nfunc F(){ gpu(family=\"metal\", name=\"m\", n=2){ x } gpu(family=\"opencl\", name=\"o\", n=3){ y } }\n"
    f := &source.File{Name: "u.ami", Content: src}
    af, errs := New(f).ParseFileCollect()
    if af == nil || len(errs) > 0 { t.Fatalf("parse errors: %+v", errs) }
    var count int
    var fams []string
    for _, d := range af.Decls {
        fn, ok := d.(*ast2.FuncDecl)
        if !ok || fn.Name != "F" || fn.Body == nil { continue }
        for _, st := range fn.Body.Stmts {
            if g, ok := st.(*ast2.GPUBlockStmt); ok {
                count++
                // find family value from attrs
                fam := ""
                for _, a := range g.Attrs {
                    if len(a.Text) >= 7 && a.Text[:7] == "family=" { fam = a.Text[7:] }
                }
                fams = append(fams, fam)
            }
        }
    }
    if count != 2 { t.Fatalf("expected 2 gpu blocks, got %d", count) }
    if len(fams) != 2 || fams[0] == fams[1] { t.Fatalf("unexpected families: %v", fams) }
}

