package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testReturnTypes_BoolFromComparison_OK(t *testing.T) {
	code := "package app\nfunc F() (bool) { return 1 == 2 }\n"
	f := (&source.FileSet{}).AddFile("rt_bool_cmp.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeReturnTypes(af)
	if len(ds) != 0 {
		t.Fatalf("unexpected diags: %+v", ds)
	}
}

func testReturnTypes_BoolFromLogical_OK(t *testing.T) {
	code := "package app\nfunc F() (bool) { return ! (1 < 2) || (3 >= 4) }\n"
	f := (&source.FileSet{}).AddFile("rt_bool_logic.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeReturnTypes(af)
	if len(ds) != 0 {
		t.Fatalf("unexpected diags: %+v", ds)
	}
}
