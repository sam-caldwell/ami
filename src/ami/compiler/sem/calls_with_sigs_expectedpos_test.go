package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Verify AnalyzeCallsWithSigs attaches expectedPos for local callees.
func TestCallsWithSigs_ExpectedPos_IncludesParamTypeLocation(t *testing.T) {
    src := "package app\n" +
        "func Callee(a string, b int) {}\n" +
        "func F(){ Callee(\"x\", \"y\") }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    params := map[string][]string{"Callee": {"string", "int"}}
    results := map[string][]string{"Callee": {}}
    ds := AnalyzeCallsWithSigs(af, params, results, nil, nil)
    if len(ds) == 0 { t.Fatalf("expected diagnostics, got none") }
    // Find the mismatch for arg index 1 (second arg), and verify expectedPos is present
    found := false
    for _, d := range ds {
        if d.Code == "E_CALL_ARG_TYPE_MISMATCH" && d.Data != nil {
            // ensure argIndex==1
            var idx int
            if v, ok := d.Data["argIndex"].(int); ok { idx = v } else if vf, ok := d.Data["argIndex"].(float64); ok { idx = int(vf) }
            if idx != 1 { continue }
            if _, ok := d.Data["expectedPos"]; !ok {
                t.Fatalf("expectedPos missing in diag data: %+v", d)
            }
            found = true
        }
    }
    if !found { t.Fatalf("missing E_CALL_ARG_TYPE_MISMATCH with expectedPos for arg1: %+v", ds) }
}
