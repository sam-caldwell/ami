package merge

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestOperator_FlushWindowExcess_EmitsHead(t *testing.T) {
    p := Plan{Window: 1}
    op := NewOperator(p)
    _ = op.Push(ev.Event{Payload: map[string]any{"v": 1}})
    _ = op.Push(ev.Event{Payload: map[string]any{"v": 2}})
    out := op.FlushWindowExcess()
    if len(out) != 1 { t.Fatalf("expected 1 flushed, got %d", len(out)) }
}

