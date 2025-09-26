package edge

import "testing"

// Verify Pipeline bridging queue honors dropNewest semantics.
func TestPipeline_DropNewestBehavior(t *testing.T) {
    p := &Pipeline{UpstreamName: "ing", MaxCapacity: 2, Backpressure: BackpressureDropNewest}
    _ = p.Push(Event{Payload: 1})
    _ = p.Push(Event{Payload: 2})
    if c := p.Count(); c != 2 {
        t.Fatalf("count after fill = %d, want 2", c)
    }
    if err := p.Push(Event{Payload: 3}); err != nil {
        t.Fatalf("unexpected error on dropNewest push: %v", err)
    }
    if c := p.Count(); c != 2 {
        t.Fatalf("count after dropNewest push = %d, want 2", c)
    }
    if ev, ok := p.Pop(); !ok || ev.Payload.(int) != 1 {
        t.Fatalf("pop1 got=%v ok=%v, want 1 true", ev.Payload, ok)
    }
    if ev, ok := p.Pop(); !ok || ev.Payload.(int) != 2 {
        t.Fatalf("pop2 got=%v ok=%v, want 2 true", ev.Payload, ok)
    }
}

