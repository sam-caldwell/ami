package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestPipelineSemantics_SelfEdge_Error(t *testing.T) {
    code := "package app\npipeline P(){ ingress; A; egress; A -> A; }\n"
    f := (&source.FileSet{}).AddFile("self.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineSemantics(af)
    has := false
    for _, d := range ds { if d.Code == "E_PIPELINE_SELF_EDGE" { has = true } }
    if !has { t.Fatalf("expected E_PIPELINE_SELF_EDGE; got %+v", ds) }
}

func TestPipelineSemantics_DuplicateEdge_Warn(t *testing.T) {
    code := "package app\npipeline P(){ ingress; A; B; egress; A -> B; A -> B; }\n"
    f := (&source.FileSet{}).AddFile("dup.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineSemantics(af)
    has := false
    for _, d := range ds { if d.Code == "W_PIPELINE_DUP_EDGE" { has = true } }
    if !has { t.Fatalf("expected W_PIPELINE_DUP_EDGE; got %+v", ds) }
}

