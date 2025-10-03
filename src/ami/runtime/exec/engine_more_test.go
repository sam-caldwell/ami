package exec

import (
    "context"
    "testing"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestEngine_RunMerge_ErrOnUninitialized(t *testing.T) {
    var e *Engine
    if _, err := e.RunMerge(context.Background(), ir.MergePlan{}, make(chan ev.Event)); err == nil { t.Fatalf("expected error on nil engine") }
    e = &Engine{}
    if _, err := e.RunMerge(context.Background(), ir.MergePlan{}, make(chan ev.Event)); err == nil { t.Fatalf("expected error on nil pool") }
}

func TestEngine_RunMerge_SimpleFlow(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{Schedule: "fifo", Concurrency: 1})
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    in := make(chan ev.Event, 4)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    plan := ir.MergePlan{} // pass-through
    out, err := eng.RunMerge(ctx, plan, in)
    if err != nil { t.Fatalf("runmerge: %v", err) }
    in <- ev.Event{Payload: map[string]any{"x": 1}}
    in <- ev.Event{Payload: map[string]any{"x": 2}}
    time.Sleep(5*time.Millisecond)
    cancel()
    // Drain
    count := 0
    for range out { count++ }
    if count == 0 { t.Fatalf("expected some output events") }
}

