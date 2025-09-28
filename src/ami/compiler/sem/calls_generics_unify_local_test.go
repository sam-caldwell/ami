package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Using AnalyzeCalls (local funcs) ensure generic compatibility on Owned<T> unifies.
func TestAnalyzeCalls_GenericOwned_Unifies_Local(t *testing.T) {
    code := "package app\nfunc H(a Owned<T>){}\nfunc F(){ var y Owned<int>; H(y) }\n"
    f := (&source.FileSet{}).AddFile("cg.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeCalls(af)
    for _, d := range ds { if d.Code == "E_CALL_ARG_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

