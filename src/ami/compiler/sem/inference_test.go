package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestTypeInference_M1_LocalsAndBinary(t *testing.T) {
    src := "package app\nfunc F(a int){ var x int; x = 1+2; x = a+2 }"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeTypeInference(af)
    if len(ds) != 0 { t.Fatalf("unexpected diags: %+v", ds) }
}

func TestTypeInference_Mismatch_OnAssign(t *testing.T) {
    src := "package app\nfunc F(){ var s string; s = 1+2 }"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeTypeInference(af)
    if len(ds) == 0 { t.Fatalf("expected mismatch diag") }
    found := false
    for _, d := range ds { if d.Code == "E_TYPE_MISMATCH" { found = true } }
    if !found { t.Fatalf("missing E_TYPE_MISMATCH: %+v", ds) }
}

func TestAmbiguity_EmptyContainers(t *testing.T) {
    src := "package app\nfunc F(){ var a = slice<any>{}; var b = set<any>{}; var c = map<any,any>{} }"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeAmbiguity(af)
    if len(ds) < 3 { t.Fatalf("expected ambiguity diags; got %d: %+v", len(ds), ds) }
}

