package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeEventTypeFlow_Mismatch(t *testing.T) {
    code := "package app\npipeline P() { A type(\"X\"); B type(\"Y\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ev.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    has := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { has = true } }
    if !has { t.Fatalf("expected E_EVENT_TYPE_FLOW, got %v", ds) }
}
