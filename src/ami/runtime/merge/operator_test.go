package merge

import (
	ev "github.com/sam-caldwell/ami/src/schemas/events"
	"testing"
)

func testMerge_Sort_Stable(t *testing.T) {
	var p Plan
	p.Buffer.Capacity = 10
	p.Buffer.Policy = "dropNewest"
	p.Stable = true
	p.Sort = []SortKey{{Field: "k", Order: "asc"}}
	op := NewOperator(p)
	e1 := ev.Event{Payload: map[string]any{"k": 1, "id": "a"}}
	e2 := ev.Event{Payload: map[string]any{"k": 1, "id": "b"}}
	if err := op.Push(e1); err != nil {
		t.Fatal(err)
	}
	if err := op.Push(e2); err != nil {
		t.Fatal(err)
	}
	o1, ok := op.Pop()
	if !ok {
		t.Fatal("no pop1")
	}
	o2, ok := op.Pop()
	if !ok {
		t.Fatal("no pop2")
	}
	if o1.Payload.(map[string]any)["id"] != "a" || o2.Payload.(map[string]any)["id"] != "b" {
		t.Fatalf("stable sort order broken: %v %v", o1, o2)
	}
}

func testMerge_Dedup(t *testing.T) {
	var p Plan
	p.Buffer.Capacity = 10
	p.Key = "id"
	p.Dedup.Field = "" // use Key
	op := NewOperator(p)
	e1 := ev.Event{Payload: map[string]any{"id": "x"}}
	e2 := ev.Event{Payload: map[string]any{"id": "x"}}
	_ = op.Push(e1)
	_ = op.Push(e2)
	if _, ok := op.Pop(); !ok {
		t.Fatal("expected first")
	}
	if _, ok := op.Pop(); ok {
		t.Fatal("expected dedup drop")
	}
}

func testMerge_Backpressure_DropNewest(t *testing.T) {
	var p Plan
	p.Buffer.Capacity = 1
	p.Buffer.Policy = "dropNewest"
	op := NewOperator(p)
	_ = op.Push(ev.Event{Payload: map[string]any{"x": 1}})
	_ = op.Push(ev.Event{Payload: map[string]any{"x": 2}}) // should drop
	out, ok := op.Pop()
	if !ok {
		t.Fatal("no out")
	}
	if out.Payload.(map[string]any)["x"].(int) != 1 {
		t.Fatalf("got %v", out)
	}
	enq, emit, drop, exp := op.Stats()
	if enq != 1 || emit != 1 || drop < 1 || exp != 0 {
		t.Fatalf("stats unexpected enq=%d emit=%d drop=%d exp=%d", enq, emit, drop, exp)
	}
}
