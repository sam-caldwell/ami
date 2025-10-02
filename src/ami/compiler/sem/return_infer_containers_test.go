package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testReturnInfer_SliceLiteral_NoDecl_OK(t *testing.T) {
	code := "package app\nfunc Ret(){ return slice<int>{1,2} }\n"
	f := (&source.FileSet{}).AddFile("t_ret_slice.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeReturnInference(af)
	// No E_TYPE_UNINFERRED expected since deduced element type is int
	for _, d := range ds {
		if d.Code == "E_TYPE_UNINFERRED" {
			t.Fatalf("unexpected uninferred: %+v", ds)
		}
	}
}

func testReturnInfer_MapLiteral_NoDecl_OK(t *testing.T) {
	code := "package app\nfunc Ret(){ return map<string,int>{\"a\":1, \"b\":2} }\n"
	f := (&source.FileSet{}).AddFile("t_ret_map.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeReturnInference(af)
	for _, d := range ds {
		if d.Code == "E_TYPE_UNINFERRED" {
			t.Fatalf("unexpected uninferred: %+v", ds)
		}
	}
}
