package buffer

import "testing"

func TestFIFO_BlockPolicy(t *testing.T) {
    q := &FIFOQueue{MaxCapacity: 2, Backpressure: "block"}
    if err := q.Push(1); err != nil { t.Fatalf("push1: %v", err) }
    if err := q.Push(2); err != nil { t.Fatalf("push2: %v", err) }
    if err := q.Push(3); err == nil { t.Fatalf("expected ErrFull") }
    if q.Len() != 2 { t.Fatalf("len: %d", q.Len()) }
    p, pop, drop, full := q.Counters()
    if p != 3 || pop != 0 || drop != 0 || full != 1 { t.Fatalf("counters: %d %d %d %d", p, pop, drop, full) }
}

func TestFIFO_DropOldestPolicy(t *testing.T) {
    q := &FIFOQueue{MaxCapacity: 2, Backpressure: "dropOldest"}
    _ = q.Push(1); _ = q.Push(2)
    if err := q.Push(3); err != nil { t.Fatalf("push overflow: %v", err) }
    if q.Len() != 2 { t.Fatalf("len: %d", q.Len()) }
    // Oldest (1) should have been dropped
    if v, ok := q.Pop(); !ok || v.(int) != 2 { t.Fatalf("pop got %v ok=%v", v, ok) }
    _, _, drop, _ := q.Counters(); if drop != 1 { t.Fatalf("dropN: %d", drop) }
}

func TestFIFO_DropNewestPolicy(t *testing.T) {
    q := &FIFOQueue{MaxCapacity: 1, Backpressure: "dropNewest"}
    _ = q.Push(1)
    if err := q.Push(2); err != nil { t.Fatalf("push overflow: %v", err) }
    if q.Len() != 1 { t.Fatalf("len: %d", q.Len()) }
    if v, ok := q.Pop(); !ok || v.(int) != 1 { t.Fatalf("pop got %v ok=%v", v, ok) }
}
