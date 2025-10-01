package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure attribute argument tokenizer preserves commas inside string literals
// such as Union type payloads for Event<...>.
func TestParser_AttrArgs_StringWithComma(t *testing.T) {
	f := &source.File{Name: "u.ami", Content: "package app\npipeline P(){ A type(\"Event<Union<int,int64>>\"); egress }\n"}
	p := New(f)
	af, _ := p.ParseFile()
	var got string
	for _, d := range af.Decls {
		if pd, ok := d.(*ast.PipelineDecl); ok {
			for _, s := range pd.Stmts {
				if st, ok := s.(*ast.StepStmt); ok && st.Name == "A" {
					for _, at := range st.Attrs {
						if at.Name == "type" && len(at.Args) > 0 {
							got = at.Args[0].Text
						}
					}
				}
			}
		}
	}
	if got != "Event<Union<int,int64>>" {
		t.Fatalf("attr arg tokenizer failed; got %q", got)
	}
}
