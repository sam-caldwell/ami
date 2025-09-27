package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestNameResolution_UnresolvedIdent(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "package app\nfunc F(){ return y }"}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeNameResolution(af)
    if len(ds) == 0 { t.Fatalf("expected unresolved ident diag") }
}

func TestNameResolution_ResolvedParamsAndVars(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "package app\nfunc F(y int){ var x int; x = y; return x }"}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeNameResolution(af)
    if len(ds) != 0 { t.Fatalf("unexpected diags: %+v", ds) }
}

