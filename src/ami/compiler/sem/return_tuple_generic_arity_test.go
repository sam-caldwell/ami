package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestReturnTypesWithSigs_Tuple_GenericArity_Mismatch_Single(t *testing.T) {
    code := "package app\n" +
        "func Producer() (Owned<int,string>, Error<int>) { return }\n" +
        "func F() (Owned<T>, Error<E>) { return Producer() }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    results := map[string][]string{
        "Producer": {"Owned<int,string>", "Error<int>"},
    }
    ds := AnalyzeReturnTypesWithSigs(af, results)
    found := false
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" {
            found = true
        }
    }
    if !found {
        t.Fatalf("expected E_GENERIC_ARITY_MISMATCH; got %+v", ds)
    }
}

func TestReturnTypesWithSigs_Tuple_GenericArity_Mismatch_Multi(t *testing.T) {
    code := "package app\n" +
        "func Producer() (Owned<int,string>, Error<int,string>) { return }\n" +
        "func F() (Owned<T>, Error<E>) { return Producer() }\n"
    var fs source.FileSet
    f := fs.AddFile("u2.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    results := map[string][]string{
        "Producer": {"Owned<int,string>", "Error<int,string>"},
    }
    ds := AnalyzeReturnTypesWithSigs(af, results)
    count := 0
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" { count++ }
    }
    if count != 2 {
        t.Fatalf("expected 2 E_GENERIC_ARITY_MISMATCH diags; got %d (%+v)", count, ds)
    }
}

