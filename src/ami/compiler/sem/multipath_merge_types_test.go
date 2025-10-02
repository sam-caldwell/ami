package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

// Watermark: invalid lateness unit/type should emit E_MERGE_ATTR_TYPE
func testMergeAttrTypes_Watermark_InvalidLateness(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ Collect merge.Watermark(ts, abc); egress }\n"
	f := (&source.FileSet{}).AddFile("wm1.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMultiPath(af)
	found := false
	for _, d := range ds {
		if d.Code == "E_MERGE_ATTR_TYPE" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_MERGE_ATTR_TYPE for invalid watermark lateness; got %+v", ds)
	}
}

// Window: non-integer emits E_MERGE_ATTR_TYPE; zero warns W_MERGE_WINDOW_ZERO_OR_NEGATIVE
func testMergeAttrTypes_Window_NumberChecks(t *testing.T) {
	f1 := (&source.FileSet{}).AddFile("w1.ami", "package app\npipeline P(){ Collect merge.Window(x); egress }\n")
	af1, _ := parser.New(f1).ParseFile()
	ds1 := AnalyzeMultiPath(af1)
	if !hasCodeRec(ds1, "E_MERGE_ATTR_TYPE") {
		t.Fatalf("expected E_MERGE_ATTR_TYPE; got %+v", ds1)
	}

	f2 := (&source.FileSet{}).AddFile("w2.ami", "package app\npipeline P(){ Collect merge.Window(0); egress }\n")
	af2, _ := parser.New(f2).ParseFile()
	ds2 := AnalyzeMultiPath(af2)
	if !hasCodeRec(ds2, "W_MERGE_WINDOW_ZERO_OR_NEGATIVE") {
		t.Fatalf("expected W_MERGE_WINDOW_ZERO_OR_NEGATIVE; got %+v", ds2)
	}
}

// Timeout: non-integer emits E_MERGE_ATTR_TYPE; zero emits E_MERGE_ATTR_ARGS (>0 required)
func testMergeAttrTypes_Timeout_NumberChecks(t *testing.T) {
	f1 := (&source.FileSet{}).AddFile("t1.ami", "package app\npipeline P(){ Collect merge.Timeout(abc); egress }\n")
	af1, _ := parser.New(f1).ParseFile()
	ds1 := AnalyzeMultiPath(af1)
	if !hasCodeRec(ds1, "E_MERGE_ATTR_TYPE") {
		t.Fatalf("expected E_MERGE_ATTR_TYPE; got %+v", ds1)
	}

	f2 := (&source.FileSet{}).AddFile("t2.ami", "package app\npipeline P(){ Collect merge.Timeout(0); egress }\n")
	af2, _ := parser.New(f2).ParseFile()
	ds2 := AnalyzeMultiPath(af2)
	if !hasCodeRec(ds2, "E_MERGE_ATTR_ARGS") {
		t.Fatalf("expected E_MERGE_ATTR_ARGS; got %+v", ds2)
	}
}

// Buffer: non-integer capacity emits E_MERGE_ATTR_TYPE
func testMergeAttrTypes_Buffer_CapacityType(t *testing.T) {
	f := (&source.FileSet{}).AddFile("b1.ami", "package app\npipeline P(){ Collect merge.Buffer(foo, dropNewest); egress }\n")
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMultiPath(af)
	if !hasCodeRec(ds, "E_MERGE_ATTR_TYPE") {
		t.Fatalf("expected E_MERGE_ATTR_TYPE; got %+v", ds)
	}
}
