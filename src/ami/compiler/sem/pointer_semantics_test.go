package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestPointer_DerefUnsafe_Error(t *testing.T) {
    src := `package p
func f() {
    *p
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_DEREF_UNSAFE" { found = true; break } }
    if !found { t.Fatalf("expected E_DEREF_UNSAFE; diags=%v", res.Diagnostics) }
}

func TestPointer_DerefSafeWithinNilGuard_OK(t *testing.T) {
    src := `package p
func f() {
    if p != nil { *p }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_DEREF_UNSAFE" || d.Code == "E_DEREF_OPERAND" {
            t.Fatalf("unexpected pointer diagnostic: %v", d)
        }
    }
}

func TestPointer_AddressOfLiteral_Error(t *testing.T) {
    src := `package p
func f() {
    mut { x = &"hi" }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_ADDR_OF_LITERAL" { found = true; break } }
    if !found { t.Fatalf("expected E_ADDR_OF_LITERAL; diags=%v", res.Diagnostics) }
}

func TestPointer_AddressOfNil_Error(t *testing.T) {
    src := `package p
func f() {
    mut { x = &nil }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_ADDR_OF_LITERAL" { found = true; break } }
    if !found { t.Fatalf("expected E_ADDR_OF_LITERAL; diags=%v", res.Diagnostics) }
}

func TestPointer_AddressOfIdent_OK(t *testing.T) {
    src := `package p
func f() {
    mut { x = &p }
}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_ADDR_OF_LITERAL" {
            t.Fatalf("unexpected E_ADDR_OF_LITERAL: %v", d)
        }
    }
}

