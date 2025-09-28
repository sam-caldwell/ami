package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestContainerLiteral_Return_Unify_Slice_OK(t *testing.T) {
    code := "package app\nfunc F() (slice<int>) { return slice<int>{1,2} }\n"
    f := &source.File{Name: "u.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeContainerTypes(af)
    for _, d := range ds { if d.Code == "E_RETURN_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestContainerLiteral_Return_Unify_Slice_Mismatch(t *testing.T) {
    code := "package app\nfunc F() (slice<int>) { return slice<int>{\"a\"} }\n"
    f := &source.File{Name: "u2.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeContainerTypes(af)
    var saw bool
    for _, d := range ds { if d.Code == "E_RETURN_TYPE_MISMATCH" { saw = true } }
    if !saw { t.Fatalf("expected E_RETURN_TYPE_MISMATCH; got %+v", ds) }
}

func TestContainerLiteral_Return_Unify_Map_OK(t *testing.T) {
    code := "package app\nfunc G() (map<string,int>) { return map<string,int>{\"k\":1} }\n"
    f := &source.File{Name: "u3.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeContainerTypes(af)
    for _, d := range ds { if d.Code == "E_RETURN_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestContainerLiteral_Return_Unify_Map_Mismatch(t *testing.T) {
    code := "package app\nfunc G() (map<string,int>) { return map<string,int>{\"k\":\"v\"} }\n"
    f := &source.File{Name: "u4.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeContainerTypes(af)
    var saw bool
    for _, d := range ds { if d.Code == "E_RETURN_TYPE_MISMATCH" { saw = true } }
    if !saw { t.Fatalf("expected E_RETURN_TYPE_MISMATCH; got %+v", ds) }
}

