package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func testMultiPath_OnlyOnCollect(t *testing.T) {
	code := "package app\npipeline P() { Alpha MultiPath(x); egress }\n"
	f := (&source.FileSet{}).AddFile("m1.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeMultiPath(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_MP_ONLY_COLLECT" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_MP_ONLY_COLLECT, got %v", ds)
	}
}

func testMergeAttr_UnknownAndArgsAndSortWarn(t *testing.T) {
	code := "package app\npipeline P() { Collect merge.Unknown(), merge.Key(), merge.Sort(); egress }\n"
	f := (&source.FileSet{}).AddFile("m2.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeMultiPath(af)
	var unknown, args, warn bool
	for _, d := range ds {
		if d.Code == "E_MERGE_ATTR_UNKNOWN" {
			unknown = true
		}
		if d.Code == "E_MERGE_ATTR_ARGS" {
			args = true
		}
		if d.Code == "W_MERGE_SORT_NO_FIELD" {
			warn = true
		}
	}
	if !unknown || !args || !warn {
		t.Fatalf("missing expected diags: %v", ds)
	}
}

func testMergeAttr_RequiredFields(t *testing.T) {
	code := "package app\npipeline P() { Collect merge.Key(\"\"), merge.PartitionBy(\"\"), merge.Sort(\"\"), merge.Watermark(\"\", 100); egress }\n"
	f := (&source.FileSet{}).AddFile("m3.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeMultiPath(af)
	errs := 0
	for _, d := range ds {
		if d.Code == "E_MERGE_ATTR_REQUIRED" {
			errs++
		}
	}
	if errs < 3 {
		t.Fatalf("expected required-field errors, got %d: %v", errs, ds)
	}
}

func testMergeBuffer_DropAlias_Warns(t *testing.T) {
	code := "package app\npipeline P() { Collect merge.Buffer(10, drop); egress }\n"
	f := (&source.FileSet{}).AddFile("m4.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeMultiPath(af)
	saw := false
	for _, d := range ds {
		if d.Code == "W_MERGE_BUFFER_DROP_ALIAS" {
			saw = true
		}
	}
	if !saw {
		t.Fatalf("expected drop alias warn: %v", ds)
	}
}
