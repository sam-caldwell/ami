package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeAmbiguity_MoreContexts(t *testing.T) {
    src := "package app\nfunc F(){ var a slice<int>; a = slice<any>{}; slice<any>{}; return }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeAmbiguity(af)
    if len(ds) == 0 { t.Fatalf("expected ambiguity diags") }
}

func TestAnalyzeTypeInference_WithSigs_CallReturn(t *testing.T) {
    src := "package app\nfunc G() (int){ return 1 }\nfunc F(){ var x int; x = G() }\n"
    f := &source.File{Name: "t2.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeTypeInference(af)
    if len(ds) != 0 { t.Fatalf("unexpected diags: %+v", ds) }
}

func TestAnalyzeCallsWithSigs_Errors(t *testing.T) {
    src := "package app\nfunc F(){ H(1, \"x\"); H(1,2,3) }\n"
    f := &source.File{Name: "t3.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    params := map[string][]string{"H": {"int", "int"}}
    ds := AnalyzeCallsWithSigs(af, params, nil, nil)
    // Expect at least one arity mismatch and one type mismatch
    if len(ds) < 2 { t.Fatalf("expected multiple call diags, got: %+v", ds) }
}
