package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestCalls_GenericArityMismatch_ProducesSpecificDiag(t *testing.T) {
    // H expects Owned<T>, Producer returns Owned<int,string>
    code := "package app\nfunc H(a Owned<T>){}\nfunc Producer() (Owned<int,string>) { return }\nfunc F(){ H(Producer()) }\n"
    f := (&source.FileSet{}).AddFile("arity.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    params := map[string][]string{"H": {"Owned<T>"}}
    results := map[string][]string{"H": {}, "Producer": {"Owned<int,string>"}}
    ds := AnalyzeCallsWithSigs(af, params, results, nil)
    found := false
    for _, d := range ds { if d.Code == "E_GENERIC_ARITY_MISMATCH" { found = true } }
    if !found { t.Fatalf("expected E_GENERIC_ARITY_MISMATCH; got %+v", ds) }
}
