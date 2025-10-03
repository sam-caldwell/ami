package merge

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "time"
)

func TestOperator_DedupAndExpireReset(t *testing.T) {
    p := Plan{}
    p.Dedup.Field = "id"
    p.TimeoutMs = 1
    op := NewOperator(p)
    _ = op.Push(ev.Event{Payload: map[string]any{"id": "x"}})
    // duplicate should be dropped; buffer remains with one item
    _ = op.Push(ev.Event{Payload: map[string]any{"id": "x"}})
    // After expire, duplicates allowed again: ExpireStale should clear buffer and reset seen
    time.Sleep(2*time.Millisecond)
    _ = op.ExpireStale(time.Now())
    _ = op.Push(ev.Event{Payload: map[string]any{"id": "x"}})
    if _, ok := op.Pop(); !ok { t.Fatalf("expected post-expire event") }
}
