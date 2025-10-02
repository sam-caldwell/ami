package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testInference_PropagateContainerInAssignment(t *testing.T) {
	src := "package app\nfunc F(){ var xs slice<int>; xs = slice<any>{1,2} }"
	f := &source.File{Name: "p1.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeTypeInference(af)
	for _, d := range ds {
		if d.Code == "E_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %v", ds)
		}
	}
}

func testReturnTypes_CompatibilityWithContainers(t *testing.T) {
	src := "package app\nfunc G() (slice<int>) { return slice<any>{1} }"
	f := &source.File{Name: "p2.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeReturnTypes(af)
	for _, d := range ds {
		if d.Code == "E_RETURN_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %v", ds)
		}
	}
}
