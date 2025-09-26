package ir

import (
	edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
	"testing"
)

func TestToPipelinesSchema_MapsEdgeSpec(t *testing.T) {
	m := Module{Package: "p", Unit: "u.ami", Pipelines: []PipelineIR{{
		Name:  "P",
        Steps: []StepIR{{Node: "Egress", In: &edg.Pipeline{UpstreamName: "X", MinCapacity: 1, MaxCapacity: 2, Backpressure: edg.BackpressureDropOldest, TypeName: "T"}}},
	}}}
	sch := m.ToPipelinesSchema()
	if len(sch.Pipelines) != 1 || len(sch.Pipelines[0].Steps) != 1 || sch.Pipelines[0].Steps[0].InEdge == nil {
		t.Fatalf("unexpected schema: %+v", sch)
	}
	e := sch.Pipelines[0].Steps[0].InEdge
    if e.Kind != "edge.Pipeline" || e.UpstreamName != "X" || e.MinCapacity != 1 || e.MaxCapacity != 2 || e.Backpressure != "dropOldest" || e.Type != "T" {
        t.Fatalf("edge mapping mismatch: %#v", e)
    }
}
