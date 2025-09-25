package ir

import (
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestLowerPipelines_WorkerContextState(t *testing.T) {
	f := &astpkg.File{Decls: []astpkg.Node{
		astpkg.FuncDecl{Name: "f", Params: []astpkg.Param{
			{Type: astpkg.TypeRef{Name: "Context"}},
			{Type: astpkg.TypeRef{Name: "Event", Args: []astpkg.TypeRef{{Name: "string"}}}},
			{Type: astpkg.TypeRef{Name: "State", Ptr: true}},
		}, Result: []astpkg.TypeRef{{Name: "Event", Args: []astpkg.TypeRef{{Name: "string"}}}}},
		astpkg.PipelineDecl{Name: "P", Steps: []astpkg.NodeCall{{Name: "Transform", Workers: []astpkg.WorkerRef{{Name: "f", Kind: "function"}}}}},
	}}
	m := Module{Package: "p", Unit: "u.ami"}
	m.LowerPipelines(f)
	if len(m.Pipelines) != 1 || len(m.Pipelines[0].Steps) != 1 {
		t.Fatalf("unexpected pipelines shape: %+v", m.Pipelines)
	}
	w := m.Pipelines[0].Steps[0].Workers
	if len(w) != 1 || !w[0].HasContext || !w[0].HasState {
		t.Fatalf("worker flags incorrect: %+v", w)
	}
}
