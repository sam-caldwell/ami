package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEventFlow_GenericEventTypeVarCompatible(t *testing.T) {
    code := "package app\npipeline P() { A type(\"Event<T>\"); B type(\"Event<int>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("g1.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds {
        if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %v", ds) }
    }
}

func TestEventFlow_GenericEventMismatch(t *testing.T) {
    code := "package app\npipeline P() { A type(\"Event<string>\"); B type(\"Event<int>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("g2.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    has := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { has = true } }
    if !has { t.Fatalf("expected mismatch for concrete Event types") }
}

