package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Mixed tuple: one generic arity mismatch and one plain type mismatch; summary should include both indices
// and include per-index path only for the generic mismatch position.
func TestReturn_MixedTuple_SummaryIncludesIndicesAndPaths(t *testing.T) {
    code := "package app\n" +
        "func Producer() (Owned<int,string>, Error<int>) { return }\n" +
        "func F() (Owned<T>, Error<E>) { return Producer(), 1 }\n"
    var fs source.FileSet
    f := fs.AddFile("ret_mixed_summary.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    results := map[string][]string{"Producer": {"Owned<int,string>", "Error<int>"}}
    ds := AnalyzeReturnTypesWithSigs(af, results)
    var seenSummary bool
    var indices []int
    var paths []map[string]any
    for _, d := range ds {
        if d.Code == "E_RETURN_TUPLE_MISMATCH_SUMMARY" && d.Data != nil {
            seenSummary = true
            if iv, ok := d.Data["indices"].([]int); ok {
                indices = iv
            } else if gv, ok := d.Data["indices"].([]any); ok {
                for _, x := range gv {
                    if i, ok := x.(int); ok { indices = append(indices, i) }
                    if f, ok := x.(float64); ok { indices = append(indices, int(f)) }
                }
            }
            if pv, ok := d.Data["paths"].([]map[string]any); ok { paths = pv }
        }
    }
    if !seenSummary { t.Fatalf("expected summary diag; got %+v", ds) }
    if len(indices) != 2 {
        t.Fatalf("expected two indices in summary; got %v (ds=%+v)", indices, ds)
    }
    // Paths should include exactly one entry for the generic mismatch at tupleIndex 0
    count0 := 0
    for _, e := range paths {
        var tix int
        if v, ok := e["tupleIndex"].(int); ok { tix = v } else if f, ok := e["tupleIndex"].(float64); ok { tix = int(f) }
        if tix == 0 { count0++ }
    }
    if count0 != 1 {
        t.Fatalf("expected one path entry for tupleIndex=0; got %d (paths=%+v, ds=%+v)", count0, paths, ds)
    }
}

