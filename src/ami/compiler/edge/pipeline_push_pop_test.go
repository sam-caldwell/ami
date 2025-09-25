package edge

import (
    "testing"
    "sync"
)

func TestPipeline_PushPop_FIFOOrder(t *testing.T) {
    q := &Pipeline{UpstreamName: "ing", Backpressure: BackpressureBlock}
    for i := 1; i <= 3; i++ { _ = q.Push(Event{Payload: i}) }
    for i := 1; i <= 3; i++ {
        ev, ok := q.Pop()
        if !ok || ev.Payload.(int) != i { t.Fatalf("want %d got %v ok=%v", i, ev.Payload, ok) }
    }
}

func TestPipeline_Backpressure_DropOldest(t *testing.T) {
    q := &Pipeline{UpstreamName: "ing", MaxCapacity: 2, Backpressure: BackpressureDrop}
    _ = q.Push(Event{Payload: 1})
    _ = q.Push(Event{Payload: 2})
    _ = q.Push(Event{Payload: 3})
    ev, ok := q.Pop()
    if !ok || ev.Payload.(int) != 2 { t.Fatalf("first want 2 got %v ok=%v", ev.Payload, ok) }
    ev, ok = q.Pop()
    if !ok || ev.Payload.(int) != 3 { t.Fatalf("second want 3 got %v ok=%v", ev.Payload, ok) }
}

func TestPipeline_BlockReturnsErrOnFull(t *testing.T) {
    q := &Pipeline{UpstreamName: "ing", MaxCapacity: 1, Backpressure: BackpressureBlock}
    _ = q.Push(Event{Payload: 1})
    if err := q.Push(Event{Payload: 2}); err == nil { t.Fatalf("expected ErrFull") }
}

func TestPipeline_ConcurrentPushes_CountMatches(t *testing.T) {
    const goroutines = 5
    const perG = 60
    q := &Pipeline{UpstreamName: "ing", Backpressure: BackpressureBlock}
    var wg sync.WaitGroup
    wg.Add(goroutines)
    for g := 0; g < goroutines; g++ {
        go func() {
            defer wg.Done()
            for i := 0; i < perG; i++ { _ = q.Push(Event{Payload: i}) }
        }()
    }
    wg.Wait()
    if got := q.Count(); got != goroutines*perG {
        t.Fatalf("concurrent pushes: want %d got %d", goroutines*perG, got)
    }
}
