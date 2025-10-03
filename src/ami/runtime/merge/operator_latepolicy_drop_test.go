package merge

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "time"
)

func TestOperator_LatePolicyDrop_DropsOld(t *testing.T) {
    p := Plan{}
    p.Watermark = &Watermark{Field: "ts", LatenessMs: 1}
    p.LatePolicy = "drop"
    op := NewOperator(p)
    past := time.Now().Add(-100*time.Millisecond).Format(time.RFC3339)
    // Old event should be dropped on Push
    _ = op.Push(ev.Event{Payload: map[string]any{"ts": past}})
    if _, ok := op.Pop(); ok { t.Fatalf("expected no buffered events due to drop") }
}

