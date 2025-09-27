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
        // also verify message contains index and types
        if d.Message == "" || exp == "" || act == "" {
            t.Fatalf("missing message/exp/act: %+v", d)
        }
        if !(contains(d.Message, "arg ") && contains(d.Message, exp) && contains(d.Message, act)) {
            t.Fatalf("message missing parts: %q", d.Message)
        }
        seen[idx] = [2]string{exp, act}
    }
    if e, ok := seen[0]; !ok || e[0] != "int" || e[1] != "string" { t.Fatalf("arg0 expected/actual: %v", seen[0]) }
    if e, ok := seen[1]; !ok || e[0] != "string" || e[1] != "int" { t.Fatalf("arg1 expected/actual: %v", seen[1]) }
}

// contains is a tiny helper to avoid importing strings in tests repeatedly.
func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0))) }
func indexOf(s, sub string) int {
    // naive implementation sufficient for tiny test strings
    n, m := len(s), len(sub)
    for i := 0; i+m <= n; i++ {
        if s[i:i+m] == sub { return i }
    }
    return -1
}
