package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEventFlow_CollectTypePropagatesToUntypedUpstream(t *testing.T) {
    code := "package app\npipeline P(){ A; B; Collect type(\"X\"); A -> Collect; B -> Collect; egress }\n"
    f := (&source.FileSet{}).AddFile("ct.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    // no mismatch expected because Collect defines type and upstreams can inherit
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected type flow error: %v", ds) } }
}

func TestEventFlow_CollectTypeConflictsWithTypedUpstream(t *testing.T) {
    code := "package app\npipeline P(){ A type(\"Y\"); Collect type(\"X\"); A -> Collect; egress }\n"
    f := (&source.FileSet{}).AddFile("ct2.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    has := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { has = true } }
    if !has { t.Fatalf("expected E_EVENT_TYPE_FLOW: %v", ds) }
}

