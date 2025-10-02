package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

// Debug helper test to verify step types and edges are collected as expected for the union case.
func testDebug_EventFlow_StepTypesAndEdges_UnionCase(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<string>\"); B type(\"Event<Union<int,int64>>\"); A -> B; egress }\n"
	f := (&source.FileSet{}).AddFile("dbg_ev.ami", code)
	af, _ := parser.New(f).ParseFile()
	var ta, tb string
	var haveEdge bool
	for _, d := range af.Decls {
		pd, ok := d.(*ast.PipelineDecl)
		if !ok {
			continue
		}
		for _, s := range pd.Stmts {
			switch st := s.(type) {
			case *ast.StepStmt:
				if st.Name == "A" || st.Name == "B" {
					for _, at := range st.Attrs {
						if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 {
							if st.Name == "A" {
								ta = at.Args[0].Text
							} else {
								tb = at.Args[0].Text
							}
						}
					}
				}
			case *ast.EdgeStmt:
				if st.From == "A" && st.To == "B" {
					haveEdge = true
				}
			}
		}
	}
	if ta != "Event<string>" {
		t.Fatalf("A type missing or wrong: %q", ta)
	}
	if tb != "Event<Union<int,int64>>" {
		t.Fatalf("B type missing or wrong: %q", tb)
	}
	if !haveEdge {
		t.Fatalf("missing edge A->B")
	}
}

func testDebug_AnalyzeEventTypeFlow_UnionMismatch(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<string>\"); B type(\"Event<Union<int,int64>>\"); A -> B; egress }\n"
	f := (&source.FileSet{}).AddFile("dbg_ev2.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeEventTypeFlow(af)
	found := false
	for _, d := range ds {
		if d.Code == "E_EVENT_TYPE_FLOW" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected union mismatch diagnostic; got %v", ds)
	}
}
