package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Verify AnalyzeCallsWithSigs attaches expectedPos for alias-qualified callees (e.g., l.Callee).
func TestCallsWithSigs_AliasQualified_ExpectedPos(t *testing.T) {
    src := "package app\n" +
        "import l \"lib\"\n" +
        "func F(){ l.Callee(\"x\", \"y\") }\n"
    var fs source.FileSet
    f := fs.AddFile("u_alias.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    // Provide param signatures and positions keyed by alias-qualified callee name
    params := map[string][]string{"l.Callee": {"string", "int"}}
    results := map[string][]string{"l.Callee": {}}
    // Put param types on line 3 in an imagined lib file; positions are synthetic here
    pp := map[string][]diag.Position{"l.Callee": {{Line: 3, Column: 16, Offset: 42}, {Line: 3, Column: 23, Offset: 49}}}
    ds := AnalyzeCallsWithSigs(af, params, results, pp, nil)
    if len(ds) == 0 { t.Fatalf("expected diagnostics, got none") }
    found := false
    for _, d := range ds {
        if d.Code == "E_CALL_ARG_TYPE_MISMATCH" && d.Data != nil {
            var idx int
            if v, ok := d.Data["argIndex"].(int); ok { idx = v } else if vf, ok := d.Data["argIndex"].(float64); ok { idx = int(vf) }
            if idx != 1 { continue }
            if ep, ok := d.Data["expectedPos"].(diag.Position); ok {
                if ep.Line != 3 { t.Fatalf("expected expectedPos line=3; got %d (diag=%+v)", ep.Line, d) }
                found = true
            }
        }
    }
    if !found { t.Fatalf("missing alias-qualified expectedPos diag: %+v", ds) }
}

