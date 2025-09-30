package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEdgeCoverage_EdgesWithoutIngress(t *testing.T) {
    // edges present; missing ingress
    code := "package app\npipeline P(){ A; egress; A -> egress; }\n"
    f := (&source.FileSet{}).AddFile("ec1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineSemantics(af)
    has := false
    for _, d := range ds { if d.Code == "E_EDGES_WITHOUT_INGRESS" { has = true; break } }
    if !has { t.Fatalf("expected E_EDGES_WITHOUT_INGRESS; got %+v", ds) }
}

func TestEdgeCoverage_EdgesWithoutEgress(t *testing.T) {
    // edges present; missing egress
    code := "package app\npipeline P(){ ingress; A; ingress -> A; }\n"
    f := (&source.FileSet{}).AddFile("ec2.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineSemantics(af)
    has := false
    for _, d := range ds { if d.Code == "E_EDGES_WITHOUT_EGRESS" { has = true; break } }
    if !has { t.Fatalf("expected E_EDGES_WITHOUT_EGRESS; got %+v", ds) }
}

func TestEdgeCoverage_EdgesWithoutIngressOrEgress(t *testing.T) {
    // edges present; missing both ingress and egress
    code := "package app\npipeline P(){ A; B; A -> B; }\n"
    f := (&source.FileSet{}).AddFile("ec3.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzePipelineSemantics(af)
    missingIngress := false
    missingEgress := false
    for _, d := range ds {
        if d.Code == "E_EDGES_WITHOUT_INGRESS" { missingIngress = true }
        if d.Code == "E_EDGES_WITHOUT_EGRESS" { missingEgress = true }
    }
    if !missingIngress || !missingEgress {
        t.Fatalf("expected both E_EDGES_WITHOUT_INGRESS and E_EDGES_WITHOUT_EGRESS; got %+v", ds)
    }
}
