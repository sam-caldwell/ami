package exec

import (
    "context"
    "testing"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestRunPipeline_FallbackIRCollectChain(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{Concurrency: 1, Schedule: "fifo"})
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    m := ir.Module{Package: "app"}
    mp := ir.MergePlan{Buffer: ir.BufferPlan{Capacity: 10, Policy: "block"}}
    m.Pipelines = []ir.Pipeline{{Name: "P", Collect: []ir.CollectSpec{{Step: "Collect", Merge: &mp}}}}
    in := make(chan ev.Event, 8)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    out, err := eng.RunPipeline(ctx, m, "P", in)
    if err != nil { t.Fatalf("run pipeline: %v", err) }
    // Feed a few events then cancel
    for i := 0; i < 3; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }
    time.Sleep(10*time.Millisecond)
    cancel()
    n := 0
    for range out { n++ }
    if n == 0 { t.Fatalf("expected outputs > 0") }
}

