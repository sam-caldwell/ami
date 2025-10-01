package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Local variable inference from initializer expressions: infer x:int from "var x = 1" and flag later mismatches.
func TestTypeInference_LocalVarFromInitializer_MismatchOnAssign(t *testing.T) {
    code := "package app\nfunc F(){ var x = 1; x = \"s\" }\n"
    f := &source.File{Name: "u.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeTypeInference(af)
    found := false
    for _, d := range ds { if d.Code == "E_TYPE_MISMATCH" { found = true } }
    if !found { t.Fatalf("expected E_TYPE_MISMATCH; got %+v", ds) }
}

// Local variable inference carries into call argument checks in AnalyzeCallsWithSigs.
func TestTypeInference_LocalVarFromInitializer_PassesToCallArgs(t *testing.T) {
    code := "package app\nfunc G(a int){}\nfunc F(){ var x = 1; G(x) }\n"
    f := &source.File{Name: "u.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFile()
    // collect params/results for calls-with-sigs
    params := map[string][]string{"G": {"int"}}
    results := map[string][]string{"G": {}}
    ds := AnalyzeCallsWithSigs(af, params, results, nil)
    for _, d := range ds { if d.Code == "E_CALL_ARG_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}
