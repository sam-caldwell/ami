package buffer

import "testing"

func TestLIFO_BlockPolicy(t *testing.T) {
    s := &LIFOStack{MaxCapacity: 1, Backpressure: "block"}
    if err := s.Push(1); err != nil { t.Fatalf("push1: %v", err) }
    if err := s.Push(2); err == nil { t.Fatalf("expected ErrFull") }
    if s.Len() != 1 { t.Fatalf("len: %d", s.Len()) }
}

func TestLIFO_DropOldestPolicy(t *testing.T) {
    s := &LIFOStack{MaxCapacity: 2, Backpressure: "dropOldest"}
    _ = s.Push(1); _ = s.Push(2)
    if err := s.Push(3); err != nil { t.Fatalf("push overflow: %v", err) }
    // Oldest (1) should have been dropped; LIFO pop should return 3 then 2
    if v, ok := s.Pop(); !ok || v.(int) != 3 { t.Fatalf("pop1: %v ok=%v", v, ok) }
    if v, ok := s.Pop(); !ok || v.(int) != 2 { t.Fatalf("pop2: %v ok=%v", v, ok) }
}

func TestLIFO_DropNewestPolicy(t *testing.T) {
    s := &LIFOStack{MaxCapacity: 1, Backpressure: "dropNewest"}
    _ = s.Push(1)
    if err := s.Push(2); err != nil { t.Fatalf("push overflow: %v", err) }
    if s.Len() != 1 { t.Fatalf("len: %d", s.Len()) }
    if v, ok := s.Pop(); !ok || v.(int) != 1 { t.Fatalf("pop got %v ok=%v", v, ok) }
}
