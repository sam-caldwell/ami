package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestPipelineNames_MultipleEntryPoints_OK(t *testing.T) {
    code := "package app\npipeline A() { ingress; egress }\npipeline B() { ingress; egress }"
    f := &source.File{Name: "pn1.ami", Content: code}
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineNames(af)
    if len(ds) != 0 { t.Fatalf("unexpected diags: %v", ds) }
}

func TestPipelineNames_Duplicate_Err(t *testing.T) {
    code := "package app\npipeline A() { ingress; egress }\npipeline A() { ingress; egress }"
    f := &source.File{Name: "pn2.ami", Content: code}
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineNames(af)
    has := false
    for _, d := range ds { if d.Code == "E_DUP_PIPELINE" { has = true } }
    if !has { t.Fatalf("expected E_DUP_PIPELINE: %v", ds) }
}
