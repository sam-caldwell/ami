package merge

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestOperator_FlushWindowExcess_EmitsHead(t *testing.T) {
    // With Window acting as effective capacity, pushes beyond the window
    // are dropped or blocked according to policy; default is drop.
    p := Plan{Window: 1}
    op := NewOperator(p)
    _ = op.Push(ev.Event{Payload: map[string]any{"v": 1}})
    _ = op.Push(ev.Event{Payload: map[string]any{"v": 2}})
    out := op.FlushWindowExcess()
    if len(out) != 0 { t.Fatalf("expected 0 flushed under window cap, got %d", len(out)) }
}
