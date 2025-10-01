package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure call aggregate summary includes per-arg path entries with argIndex.
func TestCalls_AggregateSummary_PathsIncludeArgIndex(t *testing.T) {
    src := "package app\n" +
        "func G(a Owned<T>, b Error<E>){}\n" +
        "func F(){ var x Owned<int,string>; var y Error<string,string>; G(x,y) }\n"
    var fs source.FileSet
    f := fs.AddFile("agg_paths.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeCalls(af)
    found := false
    for _, d := range ds {
        if d.Code != "E_CALL_ARGS_MISMATCH_SUMMARY" || d.Data == nil { continue }
        if pv, ok := d.Data["paths"].([]map[string]any); ok {
            if len(pv) >= 2 {
                // verify argIndex present and matches expected positions 0 and 1 in some order
                seen0, seen1 := false, false
                for _, e := range pv {
                    if idx, ok := e["argIndex"].(int); ok {
                        if idx == 0 { seen0 = true }
                        if idx == 1 { seen1 = true }
                    } else if f64, ok := e["argIndex"].(float64); ok {
                        if int(f64) == 0 { seen0 = true }
                        if int(f64) == 1 { seen1 = true }
                    }
                }
                if seen0 && seen1 { found = true; break }
            }
        }
    }
    if !found { t.Fatalf("expected summary paths to include argIndex entries; got %+v", ds) }
}

