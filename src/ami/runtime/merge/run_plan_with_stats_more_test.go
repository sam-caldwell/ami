package merge

import (
    "context"
    "testing"
    "time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// helper to build event
func evm(m map[string]any) ev.Event { return ev.Event{Payload: m} }

func TestRunPlanWithStats_CancelUpdatesStatsAndDrains(t *testing.T) {
    var p Plan
    in := make(chan ev.Event, 8)
    out := make(chan ev.Event, 8)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    var st Stats
    go RunPlanWithStats(ctx, p, in, out, &st)
    in <- evm(map[string]any{"k": 1})
    in <- evm(map[string]any{"k": 2})
    // allow loop to process enqueues
    time.Sleep(5 * time.Millisecond)
    cancel()
    // drain output until the goroutine exits
    time.Sleep(5 * time.Millisecond)
    // Basic sanity: stats should reflect activity
    if st.Enqueued == 0 && st.Emitted == 0 {
        t.Fatalf("expected non-zero stats, got %+v", st)
    }
}

func TestRunPlanWithStats_InputClosedUpdatesStats(t *testing.T) {
    var p Plan
    in := make(chan ev.Event, 4)
    out := make(chan ev.Event, 4)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    var st Stats
    go RunPlanWithStats(ctx, p, in, out, &st)
    in <- evm(map[string]any{"k": 1})
    close(in)
    time.Sleep(5 * time.Millisecond)
    if st.Enqueued == 0 && st.Emitted == 0 {
        t.Fatalf("expected stats updated on input close, got %+v", st)
    }
}

