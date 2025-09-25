package parser

import (
    "testing"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestExtractImports_SingleAndBlock(t *testing.T) {
    src := `package main
import "a/one"
import (
    "b/two"
    alias "c/three"
)
`
    got := ExtractImports(src)
    want := []string{"a/one", "b/two", "c/three"}
    if len(got) != len(want) {
        t.Fatalf("imports length mismatch: got %d want %d", len(got), len(want))
    }
    for i := range want {
        if got[i] != want[i] {
            t.Fatalf("imports[%d]=%q want %q", i, got[i], want[i])
        }
    }
}

func TestParseFile_PackageAndImports(t *testing.T) {
    src := `package pkg
import ("x/y")`
    p := New(src)
    f := p.ParseFile()
    if f.Package != "pkg" { t.Fatalf("package=%q want pkg", f.Package) }
    if len(f.Imports) != 1 || f.Imports[0] != "x/y" {
        t.Fatalf("imports=%v want [x/y]", f.Imports)
    }
}

func TestParseFile_FuncDeclScaffold(t *testing.T) {
    src := `package pkg
func main() { /* body */ }
`
    p := New(src)
    f := p.ParseFile()
    // find FuncDecl in Decls
    var count int
    for _, d := range f.Decls {
        if _, ok := d.(astpkg.FuncDecl); ok { count++ }
    }
    if count != 1 { t.Fatalf("expected 1 FuncDecl; got %d", count) }
}

func TestParsePipeline_Chain_DotAndArrow(t *testing.T) {
    src := `package pkg
pipeline P {
  Ingress(cfg).Transform(f).FanOut(a,b).Collect().Egress(cfg)
}
pipeline Q {
  Ingress(cfg) -> Transform(f) -> Egress(cfg)
}
pipeline R {
  Ingress(cfg).Transform(f).Egress(cfg) error { Collect().Egress(cfg) }
}`
    p := New(src)
    f := p.ParseFile()
    var ps []astpkg.PipelineDecl
    for _, d := range f.Decls {
        if pd, ok := d.(astpkg.PipelineDecl); ok { ps = append(ps, pd) }
    }
    if len(ps) != 3 { t.Fatalf("expected 3 pipelines; got %d", len(ps)) }
    if ps[0].Name != "P" || len(ps[0].Steps) != 5 { t.Fatalf("pipeline P steps=%d", len(ps[0].Steps)) }
    if ps[0].Steps[0].Name != "Ingress" || ps[0].Steps[4].Name != "Egress" { t.Fatalf("unexpected step names in P") }
    if ps[1].Name != "Q" || len(ps[1].Steps) != 3 { t.Fatalf("pipeline Q steps=%d", len(ps[1].Steps)) }
    if ps[1].Connectors[0] != "->" || ps[1].Connectors[1] != "->" { t.Fatalf("expected arrow connectors in Q") }
    if ps[2].Name != "R" || len(ps[2].ErrorSteps) == 0 { t.Fatalf("pipeline R should have error steps") }
}
