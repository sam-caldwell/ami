package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testPipelineSemantics_UnreachableFromIngress_WithDegree(t *testing.T) {
	// D -> B establishes degree, but neither reachable from ingress.
	code := "package app\npipeline P(){ ingress; A; B; D; egress; D -> B; }\n"
	f := (&source.FileSet{}).AddFile("u.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzePipelineSemantics(af)
	found := false
	for _, d := range ds {
		if d.Code == "E_PIPELINE_UNREACHABLE_FROM_INGRESS" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_PIPELINE_UNREACHABLE_FROM_INGRESS, got %+v", ds)
	}
}

func testPipelineSemantics_CannotReachEgress_WithDegree(t *testing.T) {
	// ingress -> A, but no path to egress.
	code := "package app\npipeline P(){ ingress; A; egress; ingress -> A; }\n"
	f := (&source.FileSet{}).AddFile("v.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzePipelineSemantics(af)
	found := false
	for _, d := range ds {
		if d.Code == "E_PIPELINE_CANNOT_REACH_EGRESS" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_PIPELINE_CANNOT_REACH_EGRESS, got %+v", ds)
	}
}
