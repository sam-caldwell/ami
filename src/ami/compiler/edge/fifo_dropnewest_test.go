package edge

import "testing"

// Verify FIFO dropNewest drops incoming items when full.
func TestFIFO_DropNewestBehavior(t *testing.T) {
    f := &FIFO{MaxCapacity: 2, Backpressure: BackpressureDropNewest}
    // push 1,2 to fill
    _ = f.Push(Event{Payload: 1})
    _ = f.Push(Event{Payload: 2})
    if c := f.Count(); c != 2 {
        t.Fatalf("count after fill = %d, want 2", c)
    }
    // push 3 -> dropped
    if err := f.Push(Event{Payload: 3}); err != nil {
        t.Fatalf("unexpected error on dropNewest push: %v", err)
    }
    if c := f.Count(); c != 2 {
        t.Fatalf("count after dropNewest push = %d, want 2", c)
    }
    // pop should return 1 then 2
    if ev, ok := f.Pop(); !ok || ev.Payload.(int) != 1 {
        t.Fatalf("pop1 got=%v ok=%v, want 1 true", ev.Payload, ok)
    }
    if ev, ok := f.Pop(); !ok || ev.Payload.(int) != 2 {
        t.Fatalf("pop2 got=%v ok=%v, want 2 true", ev.Payload, ok)
    }
}

