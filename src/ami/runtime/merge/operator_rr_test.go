package merge

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestOperator_Pop_RoundRobinAcrossPartitions(t *testing.T) {
    p := Plan{PartitionBy: "p"}
    op := NewOperator(p)
    _ = op.Push(ev.Event{Payload: map[string]any{"p": "A", "v": 1}})
    _ = op.Push(ev.Event{Payload: map[string]any{"p": "B", "v": 2}})
    e1, ok1 := op.Pop(); e2, ok2 := op.Pop()
    if !ok1 || !ok2 { t.Fatalf("expected two pops") }
    v1 := e1.Payload.(map[string]any)["v"].(int)
    v2 := e2.Payload.(map[string]any)["v"].(int)
    if v1 == v2 { t.Fatalf("expected round-robin across partitions") }
}

