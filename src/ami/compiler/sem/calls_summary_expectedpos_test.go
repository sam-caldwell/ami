package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

func TestCalls_Summary_IncludesExpectedPosInPaths(t *testing.T) {
    // Two mismatches to trigger summary, on a and b
    src := "package app\n" +
        // put params on same line; expectedPos should still be present
        "func Callee(a Owned<T>, b Error<E>) {}\n" +
        "func F(){ var x Owned<int,string>; var y Error<string,string>; Callee(x,y) }\n"
    var fs source.FileSet
    f := fs.AddFile("calls_summary_ep.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeCalls(af)
    for _, d := range ds {
        if d.Code != "E_CALL_ARGS_MISMATCH_SUMMARY" || d.Data == nil { continue }
        if pv, ok := d.Data["paths"].([]map[string]any); ok {
            for _, e := range pv {
                if _, ok := e["expectedPos"].(diag.Position); !ok {
                    t.Fatalf("expectedPos missing in call summary path entry: %+v (diag=%+v)", e, d)
                }
            }
            return
        }
    }
    t.Fatalf("call summary with expectedPos in path entries not found: %+v", ds)
}

