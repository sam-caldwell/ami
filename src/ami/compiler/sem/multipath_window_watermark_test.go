package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testMerge_Window_Without_Watermark_Warns(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ Collect merge.Window(10); egress }\n"
	f := (&source.FileSet{}).AddFile("mp_ww1.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMultiPath(af)
	found := false
	for _, d := range ds {
		if d.Code == "W_MERGE_WINDOW_WITHOUT_WATERMARK" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected W_MERGE_WINDOW_WITHOUT_WATERMARK; got %+v", ds)
	}
}

func testMerge_Watermark_Not_Primary_Sort_Warns(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ Collect merge.Sort(\"id\"), merge.Watermark(\"ts\", \"1s\"); egress }\n"
	f := (&source.FileSet{}).AddFile("mp_ww2.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMultiPath(af)
	found := false
	for _, d := range ds {
		if d.Code == "W_MERGE_WATERMARK_NOT_PRIMARY_SORT" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected W_MERGE_WATERMARK_NOT_PRIMARY_SORT; got %+v", ds)
	}
}
