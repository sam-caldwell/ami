package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testAnalyzeEventTypeFlow_MismatchAcrossEdge(t *testing.T) {
	src := "package app\npipeline P(){ Alpha type(\"int\"); Alpha -> Beta; Beta type(\"string\"); egress }\n"
	f := &source.File{Name: "ev1.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeEventTypeFlow(af)
	found := false
	for _, d := range ds {
		if d.Code == "E_EVENT_TYPE_FLOW" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_EVENT_TYPE_FLOW: %+v", ds)
	}
}

func testAnalyzeEventTypeFlow_MultiUpstreamMismatch(t *testing.T) {
	src := "package app\npipeline P(){ A type(\"int\"); B type(\"string\"); A -> Collect; B -> Collect; Collect; egress }\n"
	f := &source.File{Name: "ev2.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeEventTypeFlow(af)
	if len(ds) == 0 {
		t.Fatalf("expected mismatch on collect")
	}
}
