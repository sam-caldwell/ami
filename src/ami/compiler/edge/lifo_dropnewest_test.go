package edge

import "testing"

// Verify LIFO dropNewest drops incoming items when full.
func TestLIFO_DropNewestBehavior(t *testing.T) {
    l := &LIFO{MaxCapacity: 2, Backpressure: BackpressureDropNewest}
    _ = l.Push(Event{Payload: 1})
    _ = l.Push(Event{Payload: 2})
    if c := l.Count(); c != 2 {
        t.Fatalf("count after fill = %d, want 2", c)
    }
    if err := l.Push(Event{Payload: 3}); err != nil {
        t.Fatalf("unexpected error on dropNewest push: %v", err)
    }
    if c := l.Count(); c != 2 {
        t.Fatalf("count after dropNewest push = %d, want 2", c)
    }
    // LIFO: should pop 2 then 1
    if ev, ok := l.Pop(); !ok || ev.Payload.(int) != 2 {
        t.Fatalf("pop1 got=%v ok=%v, want 2 true", ev.Payload, ok)
    }
    if ev, ok := l.Pop(); !ok || ev.Payload.(int) != 1 {
        t.Fatalf("pop2 got=%v ok=%v, want 1 true", ev.Payload, ok)
    }
}

