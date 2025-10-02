package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testEdges_UnknownParam_Error(t *testing.T) {
	code := "package app\npipeline P() { Collect edge.FIFO(foo=1, min=1); egress }"
	f := &source.File{Name: "e5.ami", Content: code}
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEdges(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_EDGE_PARAM_UNKNOWN" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_EDGE_PARAM_UNKNOWN, got %v", ds)
	}
}

func testEdges_MaxNegative_Error(t *testing.T) {
	code := "package app\npipeline P() { Collect edge.FIFO(max=-1); egress }"
	f := &source.File{Name: "e6.ami", Content: code}
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEdges(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_EDGE_CAPACITY_INVALID" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_EDGE_CAPACITY_INVALID for max<0, got %v", ds)
	}
}

func testEdges_Backpressure_InvalidPolicy(t *testing.T) {
	code := "package app\npipeline P() { Collect edge.FIFO(backpressure=unknown); egress }"
	f := &source.File{Name: "e7.ami", Content: code}
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEdges(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_EDGE_BACKPRESSURE" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_EDGE_BACKPRESSURE for invalid value, got %v", ds)
	}
}
