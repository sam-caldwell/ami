package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestCallsWithSigs_Generic_Event_Unifies(t *testing.T) {
    src := "package app\nfunc G(a Event<T>){}\nfunc F(){ var x Event<string>; G(x) }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    params := map[string][]string{"G": {"Event<T>"}}
    results := map[string][]string{"G": {}}
    ds := AnalyzeCallsWithSigs(af, params, results)
    for _, d := range ds { if d.Code == "E_CALL_ARG_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestCallsWithSigs_Generic_Owned_Unifies(t *testing.T) {
    src := "package app\nfunc H(a Owned<T>){}\nfunc F(){ var y Owned<int>; H(y) }\n"
    var fs source.FileSet
    f := fs.AddFile("u2.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    params := map[string][]string{"H": {"Owned<T>"}}
    results := map[string][]string{"H": {}}
    ds := AnalyzeCallsWithSigs(af, params, results)
    for _, d := range ds { if d.Code == "E_CALL_ARG_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

