package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestCalls_Summary_IncludesParamName_Local(t *testing.T) {
    src := "package app\n" +
        "func Callee(a Owned<T>, b Error<E>) {}\n" +
        "func F(){ var x Owned<int,string>; var y Error<string,string>; Callee(x,y) }\n"
    var fs source.FileSet
    f := fs.AddFile("p.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeCalls(af)
    for _, d := range ds {
        if d.Code != "E_CALL_ARGS_MISMATCH_SUMMARY" || d.Data == nil { continue }
        if pv, ok := d.Data["paths"].([]map[string]any); ok {
            sawA, sawB := false, false
            for _, e := range pv {
                if n, ok := e["paramName"].(string); ok {
                    if n == "a" { sawA = true }
                    if n == "b" { sawB = true }
                }
            }
            if !sawA || !sawB { t.Fatalf("expected paramName for both args; got %+v", pv) }
            return
        }
    }
    t.Fatalf("call summary not found or missing paths: %+v", ds)
}

func TestCalls_Summary_IncludesParamName_WithSigs(t *testing.T) {
    src := "package app\n" +
        "func F(){ var x Owned<int,string>; var y Error<string,string>; Callee(x,y) }\n"
    var fs source.FileSet
    f := fs.AddFile("q.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    params := map[string][]string{"Callee": {"Owned<T>", "Error<E>"}}
    names := map[string][]string{"Callee": {"a", "b"}}
    ds := AnalyzeCallsWithSigs(af, params, nil, nil, names)
    for _, d := range ds {
        if d.Code != "E_CALL_ARGS_MISMATCH_SUMMARY" || d.Data == nil { continue }
        if pv, ok := d.Data["paths"].([]map[string]any); ok {
            sawA, sawB := false, false
            for _, e := range pv {
                if n, ok := e["paramName"].(string); ok {
                    if n == "a" { sawA = true }
                    if n == "b" { sawB = true }
                }
            }
            if !sawA || !sawB { t.Fatalf("expected paramName for both args; got %+v", pv) }
            return
        }
    }
    t.Fatalf("call summary not found or missing paths: %+v", ds)
}
