package merge

import (
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "testing"
)

func TestOperator_Push_BackpressureBlock(t *testing.T) {
    p := Plan{}
    p.Buffer.Capacity = 1
    p.Buffer.Policy = "block"
    op := NewOperator(p)
    if err := op.Push(ev.Event{Payload: map[string]any{"x":1}}); err != nil { t.Fatalf("push1: %v", err) }
    if err := op.Push(ev.Event{Payload: map[string]any{"x":2}}); err == nil { t.Fatalf("expected ErrBackpressure") }
}

func TestOperator_Push_DropNewest(t *testing.T) {
    p := Plan{}
    p.Buffer.Capacity = 1
    p.Buffer.Policy = "dropNewest"
    op := NewOperator(p)
    _ = op.Push(ev.Event{Payload: map[string]any{"x":1}})
    if err := op.Push(ev.Event{Payload: map[string]any{"x":2}}); err != nil { t.Fatalf("push2: %v", err) }
    if enq, _, drop, _ := op.Stats(); enq != 1 || drop != 1 { t.Fatalf("stats enq=%d drop=%d", enq, drop) }
}

