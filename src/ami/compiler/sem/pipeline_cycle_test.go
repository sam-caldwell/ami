package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestPipelineSemantics_Cycle_Detected(t *testing.T) {
    code := "package app\npipeline P(){ ingress; A; B; A -> B; B -> A; egress }\n"
    f := (&source.FileSet{}).AddFile("t.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    found := false
    for _, d := range ds { if d.Code == "E_PIPELINE_CYCLE" { found = true } }
    if !found { t.Fatalf("expected E_PIPELINE_CYCLE in diags: %+v", ds) }
}

