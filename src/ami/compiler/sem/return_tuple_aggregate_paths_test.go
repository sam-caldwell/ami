package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure return aggregate summary includes per-index path entries with tupleIndex.
func TestReturn_AggregateSummary_PathsIncludeTupleIndex(t *testing.T) {
    code := "package app\n" +
        "func P() (Owned<int,string>, Error<int,string>) { return }\n" +
        "func F() (Owned<T>, Error<E>) { return P() }\n"
    var fs source.FileSet
    f := fs.AddFile("ret_paths.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    results := map[string][]string{"P": {"Owned<int,string>", "Error<int,string>"}}
    ds := AnalyzeReturnTypesWithSigs(af, results)
    found := false
    for _, d := range ds {
        if d.Code != "E_RETURN_TUPLE_MISMATCH_SUMMARY" || d.Data == nil { continue }
        if pv, ok := d.Data["paths"].([]map[string]any); ok {
            if len(pv) >= 2 {
                seen0, seen1 := false, false
                for _, e := range pv {
                    // both index and tupleIndex should be present and equal to 0 or 1
                    var idx, tix int
                    if v, ok := e["index"].(int); ok { idx = v } else if f, ok := e["index"].(float64); ok { idx = int(f) }
                    if v, ok := e["tupleIndex"].(int); ok { tix = v } else if f, ok := e["tupleIndex"].(float64); ok { tix = int(f) }
                    if idx == tix {
                        if idx == 0 { seen0 = true }
                        if idx == 1 { seen1 = true }
                    }
                }
                if seen0 && seen1 { found = true; break }
            }
        }
    }
    if !found { t.Fatalf("expected summary paths to include tupleIndex entries; got %+v", ds) }
}

