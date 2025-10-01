package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestReturn_NestedGenericArityMismatch_SliceOwned_LocalVar(t *testing.T) {
    code := "package app\nfunc F() (slice<Owned<T>>) { var x slice<Owned<int,string>>; return x }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeReturnTypes(af)
    found := false
    for _, d := range ds { if d.Code == "E_GENERIC_ARITY_MISMATCH" { found = true } }
    if !found { t.Fatalf("expected nested E_GENERIC_ARITY_MISMATCH; got %+v", ds) }
}

