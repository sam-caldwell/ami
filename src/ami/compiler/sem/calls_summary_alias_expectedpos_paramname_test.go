package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Ensure E_CALL_ARGS_MISMATCH_SUMMARY 'paths' entries include expectedPos and paramName
// for alias-qualified calls via AnalyzeCallsWithSigs.
func TestCalls_Summary_AliasQualified_Paths_ExpectedPos_ParamName(t *testing.T) {
    src := "package app\n" +
        "import l \"lib\"\n" +
        // Force two mismatches to produce a summary
        "func F(){ var x Owned<int,string>; var y Error<string,string>; l.Callee(x,y) }\n"
    var fs source.FileSet
    f := fs.AddFile("calls_summary_alias.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    // Signature and positions for alias-qualified callee
    params := map[string][]string{"l.Callee": {"Owned<T>", "Error<E>"}}
    // Put param type positions on line 3 (synthetic) and include names
    ppos := map[string][]diag.Position{"l.Callee": {{Line: 3, Column: 16, Offset: 42}, {Line: 3, Column: 27, Offset: 53}}}
    pnames := map[string][]string{"l.Callee": {"a", "b"}}
    ds := AnalyzeCallsWithSigs(af, params, nil, ppos, pnames)
    found := false
    for _, d := range ds {
        if d.Code != "E_CALL_ARGS_MISMATCH_SUMMARY" || d.Data == nil { continue }
        // Expect 'paths' slice with entries for argIndex 0 and 1, each with expectedPos and paramName
        pv, ok := d.Data["paths"].([]map[string]any)
        if !ok || len(pv) < 2 { continue }
        seen0, seen1 := false, false
        for _, e := range pv {
            // expectedPos present
            if _, ok := e["expectedPos"].(diag.Position); !ok {
                t.Fatalf("expectedPos missing in alias summary path entry: %+v (diag=%+v)", e, d)
            }
            // paramName present
            if _, ok := e["paramName"].(string); !ok {
                t.Fatalf("paramName missing in alias summary path entry: %+v (diag=%+v)", e, d)
            }
            // mark argIndex visibility
            var idx int
            if v, ok := e["argIndex"].(int); ok { idx = v } else if vf, ok := e["argIndex"].(float64); ok { idx = int(vf) }
            if idx == 0 { seen0 = true }
            if idx == 1 { seen1 = true }
        }
        if !(seen0 && seen1) {
            t.Fatalf("expected both argIndex 0 and 1 in summary paths: %+v", pv)
        }
        found = true
        break
    }
    if !found { t.Fatalf("missing E_CALL_ARGS_MISMATCH_SUMMARY with alias-qualified paths including expectedPos and paramName: %+v", ds) }
}

