package exec

import (
    "context"
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestRunPipeline_NoPipelineFound_ReturnsInput(t *testing.T) {
    eng, _ := NewEngineFromModule(ir.Module{Concurrency: 1})
    defer eng.Close()
    in := make(chan ev.Event)
    out, err := eng.RunPipeline(context.Background(), ir.Module{Package: "app"}, "Unknown", in)
    if err != nil { t.Fatalf("err: %v", err) }
    if out != in { t.Fatalf("expected identity channel when pipeline not found") }
}

