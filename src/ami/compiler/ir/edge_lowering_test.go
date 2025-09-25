package ir

import (
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
	"testing"
)

func TestLowerPipelines_ParsesEdgeFIFOInArg(t *testing.T) {
	// Build AST with a single pipeline step that includes an in=edge.FIFO(...)
	pd := astpkg.PipelineDecl{
		Name: "P",
		Steps: []astpkg.NodeCall{{
			Name: "Egress",
			Args: []string{"in=edge.FIFO(minCapacity=10,maxCapacity=20,backpressure=block,type=[]byte)"},
		}},
	}
	f := &astpkg.File{Decls: []astpkg.Node{pd}}
	m := Module{Package: "p", Unit: "u.ami"}
	m.LowerPipelines(f)
	if len(m.Pipelines) != 1 || len(m.Pipelines[0].Steps) != 1 {
		t.Fatalf("unexpected pipelines shape: %+v", m.Pipelines)
	}
	in := m.Pipelines[0].Steps[0].In
	fifo, ok := in.(*edg.FIFO)
	if !ok {
		t.Fatalf("expected FIFO spec, got %#v", in)
	}
	if fifo.MinCapacity != 10 || fifo.MaxCapacity != 20 || fifo.Backpressure != edg.BackpressureBlock || fifo.TypeName != "[]byte" {
		t.Fatalf("unexpected FIFO fields: %#v", fifo)
	}
}

func TestLowerPipelines_ParsesEdgePipelineInArg(t *testing.T) {
	pd := astpkg.PipelineDecl{
		Name: "Q",
		Steps: []astpkg.NodeCall{{
			Name: "Transform",
			Args: []string{"in=edge.Pipeline(name=csvReaderPipeline,minCapacity=64,maxCapacity=128,backpressure=drop,type=csv.Record)"},
		}},
	}
	f := &astpkg.File{Decls: []astpkg.Node{pd}}
	m := Module{Package: "p", Unit: "u.ami"}
	m.LowerPipelines(f)
	in := m.Pipelines[0].Steps[0].In
	p, ok := in.(*edg.Pipeline)
	if !ok {
		t.Fatalf("expected Pipeline spec, got %#v", in)
	}
	if p.UpstreamName != "csvReaderPipeline" || p.MinCapacity != 64 || p.MaxCapacity != 128 || p.Backpressure != edg.BackpressureDrop || p.TypeName != "csv.Record" {
		t.Fatalf("unexpected Pipeline fields: %#v", p)
	}
}
