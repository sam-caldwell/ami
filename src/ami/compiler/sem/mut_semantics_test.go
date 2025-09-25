package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestMutability_AssignOutsideMut_Error(t *testing.T) {
    src := `package p
func f() { x = 1 }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_MUT_ASSIGN_OUTSIDE" { found = true; break } }
    if !found { t.Fatalf("expected E_MUT_ASSIGN_OUTSIDE; diags=%v", res.Diagnostics) }
}

func TestMutability_AssignInsideMut_OK(t *testing.T) {
    src := `package p
func f() { mut { x = 1 } }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_MUT_ASSIGN_OUTSIDE" { t.Fatalf("unexpected mutability error: %v", d) } }
}

func TestMutability_NestedBlocksInsideMut_OK(t *testing.T) {
    src := `package p
func f() { mut { { { x = 1 } } } }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_MUT_ASSIGN_OUTSIDE" { t.Fatalf("unexpected mutability error: %v", d) } }
}

func TestMutability_AfterMutBlock_Error(t *testing.T) {
    src := `package p
func f() { mut { } x = 1 }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_MUT_ASSIGN_OUTSIDE" { found = true; break } }
    if !found { t.Fatalf("expected E_MUT_ASSIGN_OUTSIDE; diags=%v", res.Diagnostics) }
}

