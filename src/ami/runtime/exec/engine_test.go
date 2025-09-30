package exec

import (
    "context"
    "testing"
    "time"

    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestEngine_RunMerge_SortsByKey_WithSchedule(t *testing.T) {
    m := ir.Module{Schedule: "fair", Concurrency: 1}
    eng, err := NewEngineFromModule(m)
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    mp := ir.MergePlan{Stable: true, Sort: []ir.SortKey{{Field: "k", Order: "asc"}}}
    in := make(chan ev.Event, 4)
    ctx, cancel := context.WithCancel(context.Background())
    out, err := eng.RunMerge(ctx, mp, in)
    if err != nil { t.Fatalf("run: %v", err) }
    // enqueue out of order
    in <- ev.Event{Payload: map[string]any{"k": 2}}
    in <- ev.Event{Payload: map[string]any{"k": 1}}
    // give the loop time
    time.Sleep(20 * time.Millisecond)
    cancel()
    // gather outputs
    var ks []int
    for e := range out { ks = append(ks, e.Payload.(map[string]any)["k"].(int)) }
    if len(ks) < 2 { t.Fatalf("expected sorted outputs, got %v", ks) }
    if !(ks[0] <= ks[1]) { t.Fatalf("not sorted: %v", ks) }
}

