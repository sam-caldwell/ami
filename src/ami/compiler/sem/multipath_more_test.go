package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testAnalyzeMultiPath_Timeout_And_Watermark_Lateness(t *testing.T) {
	src := "package app\npipeline P(){ Collect merge.Timeout(0), merge.Watermark(\"ts\", 0); egress }\n"
	f := &source.File{Name: "mp_more.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeMultiPath(af)
	hasTimeoutErr := false
	hasLateWarn := false
	for _, d := range ds {
		if d.Code == "E_MERGE_ATTR_ARGS" {
			hasTimeoutErr = true
		}
		if d.Code == "W_MERGE_WATERMARK_NONPOSITIVE" {
			hasLateWarn = true
		}
	}
	if !hasTimeoutErr || !hasLateWarn {
		t.Fatalf("expected timeout err and lateness warn: %+v", ds)
	}
}

func testAnalyzeMultiPath_Dedup_Conflicts_With_Key(t *testing.T) {
	src := "package app\npipeline P(){ Collect merge.Key(\"id\"), merge.Dedup(\"ts\"); egress }\n"
	f := &source.File{Name: "mp_conf.ami", Content: src}
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
		t.Fatalf("expected conflict between Dedup and Key: %+v", ds)
	}
}
