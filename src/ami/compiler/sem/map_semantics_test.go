package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestMapType_Happy(t *testing.T) {
    src := `package p
struct S { M map<string,int> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_MAP_ARITY" || d.Code == "E_MAP_KEY_TYPE_INVALID" {
            t.Fatalf("unexpected map diagnostic: %v", d)
        }
    }
}

func TestMapType_ArityError(t *testing.T) {
    src := `package p
struct S { M map<string> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_MAP_ARITY" { found = true; break } }
    if !found { t.Fatalf("expected E_MAP_ARITY; diags=%v", res.Diagnostics) }
}

func TestMapType_KeySliceInvalid(t *testing.T) {
    src := `package p
struct S { M map<[]byte,int> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_MAP_KEY_TYPE_INVALID" { found = true; break } }
    if !found { t.Fatalf("expected E_MAP_KEY_TYPE_INVALID for slice key; diags=%v", res.Diagnostics) }
}

func TestMapType_KeyMapInvalid(t *testing.T) {
    src := `package p
struct S { M map<map<int,int>,int> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_MAP_KEY_TYPE_INVALID" { found = true; break } }
    if !found { t.Fatalf("expected E_MAP_KEY_TYPE_INVALID for map key; diags=%v", res.Diagnostics) }
}

func TestMapType_KeyPointerInvalid(t *testing.T) {
    src := `package p
struct S { M map<*T,int> }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_MAP_KEY_TYPE_INVALID" { found = true; break } }
    if !found { t.Fatalf("expected E_MAP_KEY_TYPE_INVALID for pointer key; diags=%v", res.Diagnostics) }
}

