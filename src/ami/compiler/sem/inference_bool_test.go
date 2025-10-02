package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testTypeInference_Bool_From_Equality(t *testing.T) {
	src := "package app\nfunc F(){ var b bool; b = 1 == 2 }\n"
	f := &source.File{Name: "t1.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeTypeInference(af)
	for _, d := range ds {
		if d.Code == "E_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %+v", ds)
		}
	}
}

func testTypeInference_Bool_From_Logical_And_Or(t *testing.T) {
	src := "package app\nfunc F(){ var b bool; b = (1 < 2) && (2 < 3); b = (1 < 2) || (3 < 2) }\n"
	f := &source.File{Name: "t2.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeTypeInference(af)
	for _, d := range ds {
		if d.Code == "E_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %+v", ds)
		}
	}
}

func testTypeInference_Bool_Mismatch_AssignedToInt(t *testing.T) {
	src := "package app\nfunc F(){ var i int; i = 1 < 2 }\n"
	f := &source.File{Name: "t3.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeTypeInference(af)
	found := false
	for _, d := range ds {
		if d.Code == "E_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected type mismatch for assigning bool to int")
	}
}

func testTypeInference_Bool_From_UnaryNot(t *testing.T) {
	src := "package app\nfunc F(){ var b bool; b = !(1 == 1) }\n"
	f := &source.File{Name: "t4.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeTypeInference(af)
	for _, d := range ds {
		if d.Code == "E_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %+v", ds)
		}
	}
}
