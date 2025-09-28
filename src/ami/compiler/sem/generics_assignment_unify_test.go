package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestGenerics_Assignment_Unify_Owned_OK(t *testing.T) {
    code := "package app\nfunc G() (Owned<int>) { }\nfunc F(){ var x = G(); x = G() }\n"
    f := &source.File{Name: "u.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeTypeInference(af)
    for _, d := range ds { if d.Code == "E_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestGenerics_Assignment_Unify_Owned_Mismatch(t *testing.T) {
    code := "package app\nfunc G() (Owned<int>) { }\nfunc H() (Owned<string>) { }\nfunc F(){ var x = G(); x = H() }\n"
    f := &source.File{Name: "v.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeTypeInference(af)
    var saw bool
    for _, d := range ds { if d.Code == "E_TYPE_MISMATCH" { saw = true } }
    if !saw { t.Fatalf("expected E_TYPE_MISMATCH; got %+v", ds) }
}

func TestGenerics_Assignment_Unify_Event_OK(t *testing.T) {
    code := "package app\nfunc E1() (Event<string>) { }\nfunc F(){ var e = E1(); e = E1() }\n"
    f := &source.File{Name: "w.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeTypeInference(af)
    for _, d := range ds { if d.Code == "E_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestGenerics_Assignment_Unify_Event_Mismatch(t *testing.T) {
    code := "package app\nfunc E1() (Event<string>) { }\nfunc E2() (Event<int>) { }\nfunc F(){ var e = E1(); e = E2() }\n"
    f := &source.File{Name: "z.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeTypeInference(af)
    var saw bool
    for _, d := range ds { if d.Code == "E_TYPE_MISMATCH" { saw = true } }
    if !saw { t.Fatalf("expected E_TYPE_MISMATCH; got %+v", ds) }
}

