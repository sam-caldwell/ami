package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestSlice_BracketForm_Happy(t *testing.T) {
    src := `package p
struct S { A []byte, B []Event<string> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_SLICE_ARITY" {
            t.Fatalf("unexpected slice diagnostic: %v", d)
        }
    }
}

func TestSlice_GenericForm_Happy(t *testing.T) {
    src := `package p
struct S { A slice<string> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_SLICE_ARITY" {
            t.Fatalf("unexpected slice diagnostic: %v", d)
        }
    }
}

func TestSlice_GenericForm_ArityError(t *testing.T) {
    src := `package p
struct S { A slice<string,int> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_SLICE_ARITY" { found = true; break } }
    if !found { t.Fatalf("expected E_SLICE_ARITY; diags=%v", res.Diagnostics) }
}

