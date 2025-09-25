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

func TestParseFile_FuncDecl_ParamsAndResults(t *testing.T) {
    src := `package pkg
func a(x T, y []U) V { }
func b(a A) (B, C) { }
func c() { }
`
    p := New(src)
    f := p.ParseFile()
    var count int
    var fdas []astpkg.FuncDecl
    for _, d := range f.Decls { if fd, ok := d.(astpkg.FuncDecl); ok { count++; fdas = append(fdas, fd) } }
    if count != 3 { t.Fatalf("expected 3 funcs; got %d", count) }
    if fdas[0].Name != "a" || len(fdas[0].Params) != 2 || len(fdas[0].Result) != 1 { t.Fatalf("func a shape mismatch: %+v", fdas[0]) }
    if fdas[1].Name != "b" || len(fdas[1].Params) != 1 || len(fdas[1].Result) != 2 { t.Fatalf("func b shape mismatch: %+v", fdas[1]) }
    if fdas[2].Name != "c" || len(fdas[2].Params) != 0 || len(fdas[2].Result) != 0 { t.Fatalf("func c shape mismatch: %+v", fdas[2]) }
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

func TestParseEnum_SimpleAndWithValues(t *testing.T) {
    src := `package p
enum Color { Red, Green, Blue }
enum Code { OK=200, NotFound=404, Custom="X", Neg=-1 }
`
    p := New(src)
    f := p.ParseFile()
    found := 0
    for _, d := range f.Decls {
        if e, ok := d.(astpkg.EnumDecl); ok {
            found++
            if e.Name == "Color" {
                if len(e.Members) != 3 || e.Members[0].Name != "Red" || e.Members[1].Name != "Green" || e.Members[2].Name != "Blue" {
                    t.Fatalf("enum Color members mismatch: %+v", e.Members)
                }
                if e.Members[0].Value != "" || e.Members[1].Value != "" || e.Members[2].Value != "" {
                    t.Fatalf("enum Color values should be empty (auto): %+v", e.Members)
                }
            }
            if e.Name == "Code" {
                if len(e.Members) != 4 { t.Fatalf("enum Code members=%d want 4", len(e.Members)) }
                if e.Members[0].Name != "OK" || e.Members[0].Value != "200" { t.Fatalf("OK=200 mismatch: %+v", e.Members[0]) }
                if e.Members[1].Name != "NotFound" || e.Members[1].Value != "404" { t.Fatalf("NotFound=404 mismatch: %+v", e.Members[1]) }
                if e.Members[2].Name != "Custom" || e.Members[2].Value != "\"X\"" { t.Fatalf("Custom=\"X\" mismatch: %+v", e.Members[2]) }
                if e.Members[3].Name != "Neg" || e.Members[3].Value != "-1" { t.Fatalf("Neg=-1 mismatch: %+v", e.Members[3]) }
            }
        }
    }
    if found != 2 { t.Fatalf("expected 2 enums; got %d", found) }
}

func TestParseStruct_FieldsAndTypes(t *testing.T) {
    src := `package p
struct Person { Name string, Age int }
struct Box { Data []byte; Next *Node }
`
    p := New(src)
    f := p.ParseFile()
    found := 0
    for _, d := range f.Decls {
        st, ok := d.(astpkg.StructDecl)
        if !ok { continue }
        found++
        switch st.Name {
        case "Person":
            if len(st.Fields) != 2 { t.Fatalf("Person fields=%d", len(st.Fields)) }
            if st.Fields[0].Name != "Name" || st.Fields[0].Type.Name != "string" { t.Fatalf("Person.Name type mismatch: %+v", st.Fields[0]) }
            if st.Fields[1].Name != "Age" || st.Fields[1].Type.Name != "int" { t.Fatalf("Person.Age type mismatch: %+v", st.Fields[1]) }
        case "Box":
            if len(st.Fields) != 2 { t.Fatalf("Box fields=%d", len(st.Fields)) }
            if st.Fields[0].Name != "Data" || st.Fields[0].Type.Name != "byte" || !st.Fields[0].Type.Slice { t.Fatalf("Box.Data type mismatch: %+v", st.Fields[0]) }
            if st.Fields[1].Name != "Next" || st.Fields[1].Type.Name != "Node" || !st.Fields[1].Type.Ptr { t.Fatalf("Box.Next type mismatch: %+v", st.Fields[1]) }
        }
    }
    if found != 2 { t.Fatalf("expected 2 structs; got %d", found) }
}
