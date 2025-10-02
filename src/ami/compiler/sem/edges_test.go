package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testEdges_FIFO_MinMaxOrder_Err(t *testing.T) {
	code := "package app\npipeline P() { Collect edge.FIFO(min=10, max=5, backpressure=block); egress type(\"Event<int>\") }"
	f := &source.File{Name: "e1.ami", Content: code}
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEdges(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_EDGE_CAPACITY_ORDER" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_EDGE_CAPACITY_ORDER, got %v", ds)
	}
}

func testEdges_Pipeline_NotFound(t *testing.T) {
	code := "package app\npipeline P() { Alpha edge.Pipeline(name=Missing, type=\"Event<int>\"); egress }"
	f := &source.File{Name: "e2.ami", Content: code}
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEdges(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_EDGE_PIPE_NOT_FOUND" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_EDGE_PIPE_NOT_FOUND, got %v", ds)
	}
}

func testEdges_Pipeline_TypeMismatch(t *testing.T) {
	code := "package app\npipeline X() { ingress; egress type(\"Event<int>\") }\n" +
		"pipeline P() { Alpha edge.Pipeline(name=X, type=\"Event<string>\"); egress }"
	f := &source.File{Name: "e3.ami", Content: code}
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEdges(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_EDGE_PIPE_TYPE_MISMATCH" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_EDGE_PIPE_TYPE_MISMATCH, got %v", ds)
	}
}

func testEdges_Backpressure_LegacyDrop_Warns(t *testing.T) {
	code := "package app\npipeline P() { Collect edge.FIFO(backpressure=drop); egress }"
	f := &source.File{Name: "e4.ami", Content: code}
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEdges(af)
	has := false
	for _, d := range ds {
		if d.Code == "W_EDGE_BP_LEGACY_DROP" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected legacy drop warning, got %v", ds)
	}
}
