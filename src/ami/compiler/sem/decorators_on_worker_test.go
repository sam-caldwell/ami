package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestWorkers_Reference_DecoratedFunction_Disallowed(t *testing.T) {
    src := "package app\n@metrics\nfunc F(){}\npipeline P(){ ingress; Transform(F); egress }\n"
    f := &source.File{Name: "wd.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    has := false
    for _, d := range ds { if d.Code == "E_DECORATOR_ON_WORKER" { has = true } }
    if !has { t.Fatalf("expected E_DECORATOR_ON_WORKER from AnalyzeWorkers: %+v", ds) }
}
