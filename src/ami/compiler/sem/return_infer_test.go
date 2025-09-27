package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestReturnInference_Happy(t *testing.T) {
    src := "package app\nfunc H(a int){ var xs slice<int>; xs = slice<int>{1}; return xs }"
    f := &source.File{Name: "ri1.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnInference(af)
    for _, d := range ds { if d.Code != "" { t.Fatalf("unexpected diags: %v", ds) } }
}

func TestReturnInference_Uninferred(t *testing.T) {
    src := "package app\nfunc H(){ var x; return x }"
    f := &source.File{Name: "ri2.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnInference(af)
    has := false
    for _, d := range ds { if d.Code == "E_TYPE_UNINFERRED" { has = true } }
    if !has { t.Fatalf("expected E_TYPE_UNINFERRED: %v", ds) }
}

func TestReturnInference_Mismatch(t *testing.T) {
    src := "package app\nfunc H(){ return slice<int>{1}; return slice<string>{\"x\"} }"
    f := &source.File{Name: "ri3.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnInference(af)
    has := false
    for _, d := range ds { if d.Code == "E_RETURN_TYPE_MISMATCH" { has = true } }
    if !has { t.Fatalf("expected E_RETURN_TYPE_MISMATCH: %v", ds) }
}
