package merge

import (
    "context"
    "testing"
    "time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestRunPlan_WindowedOrdering(t *testing.T) {
    var p Plan
    p.Window = 2
    p.Sort = []SortKey{{Field:"k", Order:"asc"}}
    in := make(chan ev.Event, 4)
    out := make(chan ev.Event, 4)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go RunPlan(ctx, p, in, out)
    in <- E(map[string]any{"k":2})
    in <- E(map[string]any{"k":1})
    in <- E(map[string]any{"k":3})
    // Let loop process
    time.Sleep(20 * time.Millisecond)
    // Expect only 2 items due to window cap
    var got []int
    for i := 0; i < 3; i++ {
        select { case e := <-out: got = append(got, e.Payload.(map[string]any)["k"].(int)); default: }
    }
    if len(got) != 2 { t.Fatalf("expected 2 items, got %v", got) }
    if !(got[0] <= got[1]) { t.Fatalf("not sorted: %v", got) }
}

func TestRunPlan_Watermark_FlushesReady(t *testing.T) {
    var p Plan
    p.Watermark = &Watermark{Field: "ts", LatenessMs: 50}
    in := make(chan ev.Event, 4)
    out := make(chan ev.Event, 8)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go RunPlan(ctx, p, in, out)
    // send two old events and one new
    old := time.Now().Add(-time.Second).Format(time.RFC3339)
    in <- E(map[string]any{"ts": old, "id": "a"})
    in <- E(map[string]any{"ts": old, "id": "b"})
    time.Sleep(100 * time.Millisecond) // allow flush cycle
    // collect flushed outputs
    var ids []string
    for i := 0; i < 2; i++ {
        select { case e := <-out: ids = append(ids, e.Payload.(map[string]any)["id"].(string)); default: }
    }
    if len(ids) != 2 { t.Fatalf("expected flushed 2 by watermark, got %v", ids) }
}
