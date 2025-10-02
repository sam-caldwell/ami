package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testMergeAttr_DuplicateConflict_Sort(t *testing.T) {
	code := "package app\npipeline P(){ Collect merge.Sort(\"ts\"), merge.Sort(\"id\"); egress }\n"
	f := (&source.FileSet{}).AddFile("mdup1.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeMultiPath(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_MERGE_ATTR_CONFLICT" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected conflict for duplicate merge.Sort: %+v", ds)
	}
}

func testMergeAttr_DuplicateConflict_Buffer(t *testing.T) {
	code := "package app\npipeline P(){ Collect merge.Buffer(10, block), merge.Buffer(20, block); egress }\n"
	f := (&source.FileSet{}).AddFile("mdup2.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeMultiPath(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_MERGE_ATTR_CONFLICT" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected conflict for duplicate merge.Buffer: %+v", ds)
	}
}
