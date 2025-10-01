package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

func TestReturn_Summary_IncludesFunctionAndExpectedPosInPaths(t *testing.T) {
    code := "package app\n" +
        "func P() (Owned<int,string>, Error<int,string>) { return }\n" +
        // Place F on line 3 so result type positions map to line 3
        "func F() (Owned<T>, Error<E>) { return P() }\n"
    var fs source.FileSet
    f := fs.AddFile("ret_summary_fn.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    results := map[string][]string{"P": {"Owned<int,string>", "Error<int,string>"}}
    ds := AnalyzeReturnTypesWithSigs(af, results)
    var found bool
    for _, d := range ds {
        if d.Code != "E_RETURN_TUPLE_MISMATCH_SUMMARY" || d.Data == nil { continue }
        if fn, ok := d.Data["function"].(string); !ok || fn != "F" { t.Fatalf("function missing or wrong: %+v", d) }
        if pv, ok := d.Data["paths"].([]map[string]any); ok {
            // there should be two paths; ensure expectedPos present and on line 3
            if len(pv) != 2 { t.Fatalf("expected 2 path entries; got %d (%+v)", len(pv), d) }
            for _, e := range pv {
                if ep, ok := e["expectedPos"].(diag.Position); !ok || ep.Line != 3 {
                    t.Fatalf("expectedPos missing or wrong line: %+v (diag=%+v)", e, d)
                }
            }
            found = true
            break
        }
    }
    if !found { t.Fatalf("summary with function and expectedPos not found: %+v", ds) }
}

