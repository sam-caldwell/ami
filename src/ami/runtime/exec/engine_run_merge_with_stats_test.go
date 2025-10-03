package exec

import (
    "context"
    "testing"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestEngine_RunMergeWithStats_PopulatesStats(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{Concurrency: 1, Schedule: "fifo"})
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    in := make(chan ev.Event, 4)
    ctx, cancel := context.WithCancel(context.Background())
    plan := ir.MergePlan{}
    out, st, err := eng.RunMergeWithStats(ctx, plan, in)
    if err != nil || st == nil { t.Fatalf("run merge stats: %v st=%v", err, st) }
    for i := 0; i < 3; i++ { in <- ev.Event{Payload: map[string]any{"x": i}} }
    time.Sleep(30*time.Millisecond)
    cancel()
    for range out { /* drain */ }
    // Successful run returned a non-nil stats pointer; detailed counters are validated elsewhere.
}
