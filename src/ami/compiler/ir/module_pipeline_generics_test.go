package ir

import (
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestLowerPipelines_CapturesGenericPayloads(t *testing.T) {
	f := &astpkg.File{Decls: []astpkg.Node{
		astpkg.FuncDecl{Name: "xform", Params: []astpkg.Param{
			{Type: astpkg.TypeRef{Name: "Context"}},
			{Type: astpkg.TypeRef{Name: "Event", Args: []astpkg.TypeRef{{Name: "T"}}}},
			{Type: astpkg.TypeRef{Name: "State", Ptr: true}},
		}, Result: []astpkg.TypeRef{{Name: "Event", Args: []astpkg.TypeRef{{Name: "U"}}}}},
		astpkg.PipelineDecl{Name: "P", Steps: []astpkg.NodeCall{{Name: "Transform", Workers: []astpkg.WorkerRef{{Name: "xform", Kind: "function"}}}}},
	}}
	m := Module{Package: "p", Unit: "u.ami"}
	m.LowerPipelines(f)
	if len(m.Pipelines) != 1 {
		t.Fatalf("expected 1 pipeline; got %d", len(m.Pipelines))
	}
	if len(m.Pipelines[0].Steps) != 1 {
		t.Fatalf("expected 1 step; got %d", len(m.Pipelines[0].Steps))
	}
	w := m.Pipelines[0].Steps[0].Workers
	if len(w) != 1 {
		t.Fatalf("expected 1 worker; got %d", len(w))
	}
	if w[0].Input != "T" || w[0].OutputKind != "Event" || w[0].Output != "U" {
		t.Fatalf("generic payloads not captured: %+v", w[0])
	}
}
