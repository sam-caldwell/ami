package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Pointer types in Event<T> parameter should be forbidden.
func TestWorkers_EventParam_PointerForbidden(t *testing.T) {
    code := "package app\n" +
        "func F(ev Event<*T>) (Event<U>, error) { return ev, nil }\n" +
        "pipeline P(){ ingress; Transform(F); egress }\n"
    f := (&source.FileSet{}).AddFile("w_ptr_param.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    var has bool
    for _, d := range ds { if d.Code == "E_EVENT_PTR_FORBIDDEN" { has = true } }
    if !has { t.Fatalf("expected E_EVENT_PTR_FORBIDDEN for Event<*T>: %+v", ds) }
}

// Pointer types in Event<U> result should be forbidden.
func TestWorkers_EventResult_PointerForbidden(t *testing.T) {
    code := "package app\n" +
        "func F(ev Event<T>) (Event<&U>, error) { return ev, nil }\n" +
        "pipeline P(){ ingress; Transform(F); egress }\n"
    f := (&source.FileSet{}).AddFile("w_ptr_result.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    var has bool
    for _, d := range ds { if d.Code == "E_EVENT_PTR_FORBIDDEN" { has = true } }
    if !has { t.Fatalf("expected E_EVENT_PTR_FORBIDDEN for Event<&U>: %+v", ds) }
}

