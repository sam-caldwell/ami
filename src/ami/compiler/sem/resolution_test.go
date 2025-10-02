package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testNameResolution_UnresolvedIdent(t *testing.T) {
	f := &source.File{Name: "t.ami", Content: "package app\nfunc F(){ return y }"}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeNameResolution(af)
	if len(ds) == 0 {
		t.Fatalf("expected unresolved ident diag")
	}
}

func testNameResolution_ResolvedParamsAndVars(t *testing.T) {
	f := &source.File{Name: "t.ami", Content: "package app\nfunc F(y int){ var x int; x = y; return x }"}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeNameResolution(af)
	if len(ds) != 0 {
		t.Fatalf("unexpected diags: %+v", ds)
	}
}

func testNameResolution_Unresolved_NestedPositions(t *testing.T) {
	f := &source.File{Name: "t2.ami", Content: "package app\nfunc F(){ return a + (b+1) }"}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeNameResolution(af)
	if len(ds) < 1 {
		t.Fatalf("expected unresolved idents, got %d: %+v", len(ds), ds)
	}
	for _, d := range ds {
		if d.Pos == nil || d.Pos.Line <= 0 || d.Pos.Column <= 0 {
			t.Fatalf("diag missing pos: %+v", d)
		}
	}
}
