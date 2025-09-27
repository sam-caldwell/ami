package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestErrorPipeline_StartWithIngress_Err(t *testing.T) {
    src := "package app\npipeline P(){ error { ingress().egress } }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    file, _ := p.ParseFile()
    ds := AnalyzeErrorSemantics(file)
    has := false
    for _, d := range ds { if d.Code == "E_ERRPIPE_START_INVALID" { has = true } }
    if !has { t.Fatalf("expected E_ERRPIPE_START_INVALID, got %v", ds) }
}

func TestErrorPipeline_MustEndWithEgress_Err(t *testing.T) {
    src := "package app\npipeline P(){ error { Transform() } }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    file, _ := p.ParseFile()
    ds := AnalyzeErrorSemantics(file)
    has := false
    for _, d := range ds { if d.Code == "E_ERRPIPE_END_EGRESS" { has = true } }
    if !has { t.Fatalf("expected E_ERRPIPE_END_EGRESS, got %v", ds) }
}

func TestErrorPipeline_UnknownNode_Err(t *testing.T) {
    src := "package app\npipeline P(){ error { Foo().egress } }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    file, _ := p.ParseFile()
    ds := AnalyzeErrorSemantics(file)
    has := false
    for _, d := range ds { if d.Code == "E_UNKNOWN_NODE" { has = true } }
    if !has { t.Fatalf("expected E_UNKNOWN_NODE, got %v", ds) }
}

