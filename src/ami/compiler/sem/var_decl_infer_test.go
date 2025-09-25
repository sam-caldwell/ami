package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestVarDecl_InferFromBinary_OK(t *testing.T) {
    src := `package p
func f(a int, b int) int { var x = a + b; return x }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_ASSIGN_TYPE_MISMATCH" || d.Code == "E_TYPE_UNINFERRED" || d.Code == "E_TYPE_MISMATCH" || d.Code == "E_RETURN_TYPE_MISMATCH" {
            t.Fatalf("unexpected diag: %v", d)
        }
    }
}

func TestVarDecl_AnnotationAndInit_Mismatch_Error(t *testing.T) {
    src := `package p
func f() { var x int = "hi" }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_ASSIGN_TYPE_MISMATCH" { found = true; break } }
    if !found { t.Fatalf("expected E_ASSIGN_TYPE_MISMATCH; diags=%v", res.Diagnostics) }
}

func TestVarDecl_NoAnnot_NoInit_Uninferred(t *testing.T) {
    src := `package p
func f() { var x }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_TYPE_UNINFERRED" { found = true; break } }
    if !found { t.Fatalf("expected E_TYPE_UNINFERRED; diags=%v", res.Diagnostics) }
}

