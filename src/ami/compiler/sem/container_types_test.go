package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testContainerTypes_Slice_Uniform_OK(t *testing.T) {
	src := "package app\nfunc F(){ var a slice<int>; a = slice<int>{1,2,3} }\n"
	f := &source.File{Name: "t.ami", Content: src}
	p := parser.New(f)
	file, _ := p.ParseFile()
	ds := AnalyzeContainerTypes(file)
	if len(ds) != 0 {
		t.Fatalf("unexpected diags: %+v", ds)
	}
}

func testContainerTypes_Slice_Mismatch_Err(t *testing.T) {
	src := "package app\nfunc F(){ var a slice<int>; a = slice<int>{1,\"x\"} }\n"
	f := &source.File{Name: "t.ami", Content: src}
	p := parser.New(f)
	file, _ := p.ParseFile()
	ds := AnalyzeContainerTypes(file)
	if len(ds) == 0 {
		t.Fatalf("expected mismatch diag")
	}
}

func testContainerTypes_Map_KeyAndValue_OK(t *testing.T) {
	src := "package app\nfunc F(){ var m map<string,int>; m = map<string,int>{\"k\":1, \"q\":2} }\n"
	f := &source.File{Name: "t.ami", Content: src}
	p := parser.New(f)
	file, _ := p.ParseFile()
	ds := AnalyzeContainerTypes(file)
	if len(ds) != 0 {
		t.Fatalf("unexpected diags: %+v", ds)
	}
}

func testContainerTypes_Map_KeyMismatch_Err(t *testing.T) {
	src := "package app\nfunc F(){ var m map<string,int>; m = map<string,int>{\"k\":1, 2:3} }\n"
	f := &source.File{Name: "t.ami", Content: src}
	p := parser.New(f)
	file, _ := p.ParseFile()
	ds := AnalyzeContainerTypes(file)
	has := false
	for _, d := range ds {
		if d.Code == "E_TYPE_MISMATCH" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_TYPE_MISMATCH in keys, got %+v", ds)
	}
}

func testContainerTypes_Map_ValueMismatch_Err(t *testing.T) {
	src := "package app\nfunc F(){ var m map<string,int>; m = map<string,int>{\"k\":\"x\"} }\n"
	f := &source.File{Name: "t.ami", Content: src}
	p := parser.New(f)
	file, _ := p.ParseFile()
	ds := AnalyzeContainerTypes(file)
	has := false
	for _, d := range ds {
		if d.Code == "E_TYPE_MISMATCH" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_TYPE_MISMATCH in values, got %+v", ds)
	}
}
