package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestReturn_TupleAggregateSummary_Emitted(t *testing.T) {
    // Two mismatches across tuple positions
    code := "package app\n" +
        "func Producer() (Owned<int,string>, Error<int,string>) { return }\n" +
        "func F() (Owned<T>, Error<E>) { return Producer() }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    results := map[string][]string{"Producer": {"Owned<int,string>", "Error<int,string>"}}
    ds := AnalyzeReturnTypesWithSigs(af, results)
    per := 0
    summary := 0
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" { per++ }
        if d.Code == "E_RETURN_TUPLE_MISMATCH_SUMMARY" { summary++ }
    }
    if per < 2 || summary != 1 {
        t.Fatalf("expected >=2 per-element and 1 summary; got per=%d summary=%d (%+v)", per, summary, ds)
    }
}

