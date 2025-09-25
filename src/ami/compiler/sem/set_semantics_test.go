package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestSetType_Happy(t *testing.T) {
    src := `package p
struct S { M set<string> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_SET_ARITY" || d.Code == "E_SET_ELEM_TYPE_INVALID" {
            t.Fatalf("unexpected set diagnostic: %v", d)
        }
    }
}

func TestSetType_ArityError(t *testing.T) {
    src := `package p
struct S { M set<string,int> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_SET_ARITY" { found = true; break } }
    if !found { t.Fatalf("expected E_SET_ARITY; diags=%v", res.Diagnostics) }
}

func TestSetType_ElemSliceInvalid(t *testing.T) {
    src := `package p
struct S { M set<[]byte> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_SET_ELEM_TYPE_INVALID" { found = true; break } }
    if !found { t.Fatalf("expected E_SET_ELEM_TYPE_INVALID for slice element; diags=%v", res.Diagnostics) }
}

func TestSetType_ElemMapInvalid(t *testing.T) {
    src := `package p
struct S { M set<map<int,int>> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_SET_ELEM_TYPE_INVALID" { found = true; break } }
    if !found { t.Fatalf("expected E_SET_ELEM_TYPE_INVALID for map element; diags=%v", res.Diagnostics) }
}

func TestSetType_ElemPointerInvalid(t *testing.T) {
    src := `package p
struct S { M set<*T> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_SET_ELEM_TYPE_INVALID" { found = true; break } }
    if !found { t.Fatalf("expected E_SET_ELEM_TYPE_INVALID for pointer element; diags=%v", res.Diagnostics) }
}

