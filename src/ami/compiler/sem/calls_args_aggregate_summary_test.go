package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestCalls_AggregateSummary_Emitted(t *testing.T) {
    // Callee expects (int,string); call provides (string,int) → two mismatches
    src := "package app\nfunc Callee(a int, b string) {}\nfunc F(){ Callee(\"x\", 1) }\n"
    var fs source.FileSet
    f := fs.AddFile("t.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeCalls(af)
    per := 0
    summary := 0
    for _, d := range ds {
        if d.Code == "E_CALL_ARG_TYPE_MISMATCH" { per++ }
        if d.Code == "E_CALL_ARGS_MISMATCH_SUMMARY" { summary++ }
    }
    if per < 2 || summary != 1 {
        t.Fatalf("expected >=2 per-arg mismatches and 1 summary; got per=%d summary=%d (%+v)", per, summary, ds)
    }
}

