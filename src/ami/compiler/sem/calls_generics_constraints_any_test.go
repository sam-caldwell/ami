package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure generic constraints with any: Owned<any> parameter accepts Owned<int> at call site.
func TestCalls_GenericConstraint_Any_Owned(t *testing.T) {
    code := "package app\nfunc H(a Owned<any>){}\nfunc F(){ var y Owned<int>; H(y) }\n"
    f := (&source.FileSet{}).AddFile("g_any.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeCalls(af)
    for _, d := range ds { if d.Code == "E_CALL_ARG_TYPE_MISMATCH" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

