package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestCalls_MultipleArgMismatches_ReportEachWithIndex(t *testing.T) {
    // Callee expects (int,string,int); call passes (string,int,string) → mismatches at 0 and 2.
    src := "package app\nfunc Callee(a int, b string, c int) {}\nfunc F(){ Callee(\"x\", 2, \"z\") }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeCalls(af)
    var idxs []int
    for _, d := range ds {
        if d.Code == "E_CALL_ARG_TYPE_MISMATCH" && d.Data != nil {
            if v, ok := d.Data["argIndex"].(int); ok { idxs = append(idxs, v); continue }
            if vf, ok := d.Data["argIndex"].(float64); ok { idxs = append(idxs, int(vf)); continue }
        }
    }
    if len(idxs) != 3 { t.Fatalf("expected three arg mismatches, got %v", idxs) }
    seen := map[int]bool{}
    for _, i := range idxs { seen[i] = true }
    if !seen[0] || !seen[1] || !seen[2] { t.Fatalf("unexpected mismatch indexes: %v", idxs) }
}
