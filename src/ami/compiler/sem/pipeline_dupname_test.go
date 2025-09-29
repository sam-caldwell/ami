package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestPipelineSemantics_DuplicatePipelineName_Error(t *testing.T) {
    code := "package app\npipeline P(){ ingress; egress }\npipeline P(){ ingress; egress }\n"
    f := (&source.FileSet{}).AddFile("dupname.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineSemantics(af)
    has := false
    for _, d := range ds { if d.Code == "E_DUP_PIPELINE" { has = true } }
    if !has { t.Fatalf("expected E_DUP_PIPELINE; got %+v", ds) }
}

