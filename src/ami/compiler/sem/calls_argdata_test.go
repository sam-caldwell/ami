package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestCalls_ArgTypeMismatch_DataExpectedActual(t *testing.T) {
    // Callee expects (int,string); call provides (string,int) → mismatches at both 0 and 1.
    src := "package app\nfunc Callee(a int, b string) {}\nfunc F(){ Callee(\"x\", 1) }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeCalls(af)
    if len(ds) == 0 { t.Fatalf("expected diagnostics") }
    seen := map[int][2]string{}
    for _, d := range ds {
        if d.Code != "E_CALL_ARG_TYPE_MISMATCH" || d.Data == nil { continue }
        var idx int
        if v, ok := d.Data["argIndex"].(int); ok { idx = v } else if vf, ok := d.Data["argIndex"].(float64); ok { idx = int(vf) }
        exp, _ := d.Data["expected"].(string)
        act, _ := d.Data["actual"].(string)
        seen[idx] = [2]string{exp, act}
    }
    if e, ok := seen[0]; !ok || e[0] != "int" || e[1] != "string" { t.Fatalf("arg0 expected/actual: %v", seen[0]) }
    if e, ok := seen[1]; !ok || e[0] != "string" || e[1] != "int" { t.Fatalf("arg1 expected/actual: %v", seen[1]) }
}

