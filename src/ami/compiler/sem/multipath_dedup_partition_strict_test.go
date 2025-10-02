package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testMerge_DedupFieldUnderPartition_WarnsByDefault(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ Collect merge.PartitionBy(\"p\"), merge.Dedup(\"ts\"); egress }\n"
	f := (&source.FileSet{}).AddFile("mp_dfp_warn.ami", code)
	af, _ := parser.New(f).ParseFile()
	StrictDedupUnderPartition = false
	ds := AnalyzeMultiPath(af)
	found := false
	for _, d := range ds {
		if d.Code == "W_MERGE_DEDUP_FIELD_WITHOUT_KEY_UNDER_PARTITION" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected W_MERGE_DEDUP_FIELD_WITHOUT_KEY_UNDER_PARTITION; got %+v", ds)
	}
}

func testMerge_DedupFieldUnderPartition_Errors_WhenStrict(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ Collect merge.PartitionBy(\"p\"), merge.Dedup(\"ts\"); egress }\n"
	f := (&source.FileSet{}).AddFile("mp_dfp_err.ami", code)
	af, _ := parser.New(f).ParseFile()
	StrictDedupUnderPartition = true
	ds := AnalyzeMultiPath(af)
	StrictDedupUnderPartition = false
	found := false
	for _, d := range ds {
		if d.Code == "E_MERGE_DEDUP_FIELD_WITHOUT_KEY_UNDER_PARTITION" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_MERGE_DEDUP_FIELD_WITHOUT_KEY_UNDER_PARTITION; got %+v", ds)
	}
}
