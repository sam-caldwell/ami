package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

// When both merge.Key and merge.Sort are specified, require primary sort equals key.
func testAnalyzeMultiPath_SortPrimaryMustEqualKey_Err(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ Collect merge.Key(\"id\"), merge.Sort(\"ts\"); egress }\n"
	f := (&source.FileSet{}).AddFile("mp_key_sort_warn.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMultiPath(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_MERGE_SORT_PRIMARY_NEQ_KEY" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_MERGE_SORT_PRIMARY_NEQ_KEY, got %+v", ds)
	}
}

// When only PartitionBy is present, require primary sort equals partition field.
func testAnalyzeMultiPath_SortPrimaryMustEqualPartition_Err(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ Collect merge.PartitionBy(\"p\"), merge.Sort(\"ts\"); egress }\n"
	f := (&source.FileSet{}).AddFile("mp_part_sort_err.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMultiPath(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_MERGE_SORT_PRIMARY_NEQ_PARTITION" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_MERGE_SORT_PRIMARY_NEQ_PARTITION, got %+v", ds)
	}
}
