package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestCalls_NestedGenericArityMismatch_SliceOwned_LocalVar(t *testing.T) {
    src := "package app\nfunc H(a slice<Owned<T>>){ }\nfunc F(){ var y slice<Owned<int,string>>; H(y) }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeCalls(af)
    found := false
    for _, d := range ds { if d.Code == "E_GENERIC_ARITY_MISMATCH" { found = true } }
    if !found { t.Fatalf("expected nested E_GENERIC_ARITY_MISMATCH; got %+v", ds) }
}

