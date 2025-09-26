package merge

import "testing"

func TestBuffer_BlockPolicy_ErrFull(t *testing.T) {
    o := New(Config{BufferCapacity: 1, BufferBackpressure: "block"})
    if err := o.Push(1); err != nil { t.Fatalf("push1: %v", err) }
    if err := o.Push(2); err == nil { t.Fatalf("expected ErrFull") }
}

func TestBuffer_DropOldest_Policy(t *testing.T) {
    o := New(Config{BufferCapacity: 2, BufferBackpressure: "dropOldest"})
    _ = o.Push(1)
    _ = o.Push(2)
    _ = o.Push(3) // should evict 1, keep 2,3
    if v, ok := o.Pop(); !ok || v.(int) != 2 { t.Fatalf("want 2, got %v ok=%v", v, ok) }
    if v, ok := o.Pop(); !ok || v.(int) != 3 { t.Fatalf("want 3, got %v ok=%v", v, ok) }
}

func TestBuffer_DropNewest_Policy(t *testing.T) {
    o := New(Config{BufferCapacity: 2, BufferBackpressure: "dropNewest"})
    _ = o.Push(1)
    _ = o.Push(2)
    _ = o.Push(3) // drop incoming; keep 1,2
    if v, ok := o.Pop(); !ok || v.(int) != 1 { t.Fatalf("want 1, got %v ok=%v", v, ok) }
    if v, ok := o.Pop(); !ok || v.(int) != 2 { t.Fatalf("want 2, got %v ok=%v", v, ok) }
}

