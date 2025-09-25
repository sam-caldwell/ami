package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestImperative_AssignTypeMismatch_Error(t *testing.T) {
    src := `package p
func f(a int, b string) {
    mut { a = b }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_ASSIGN_TYPE_MISMATCH" { found = true; break } }
    if !found { t.Fatalf("expected E_ASSIGN_TYPE_MISMATCH; diags=%v", res.Diagnostics) }
}

func TestImperative_AssignPointerDeref_Mismatch_Error(t *testing.T) {
    src := `package p
func f(b string, p *int) {
    if p != nil { mut { *p = b } }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_ASSIGN_TYPE_MISMATCH" { found = true; break } }
    if !found { t.Fatalf("expected E_ASSIGN_TYPE_MISMATCH; diags=%v", res.Diagnostics) }
}

func TestImperative_AssignPointerDeref_OK(t *testing.T) {
    src := `package p
func f(a int, p *int) {
    if p != nil { mut { *p = a } }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_ASSIGN_TYPE_MISMATCH" || d.Code == "E_DEREF_TYPE" || d.Code == "E_DEREF_UNSAFE" {
            t.Fatalf("unexpected diagnostic: %v", d)
        }
    }
}

func TestImperative_AssignAddressOf_OK(t *testing.T) {
    src := `package p
func f(a int, p *int) {
    mut { p = &a }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_ASSIGN_TYPE_MISMATCH" || d.Code == "E_ADDR_OF_LITERAL" {
            t.Fatalf("unexpected diagnostic: %v", d)
        }
    }
}

