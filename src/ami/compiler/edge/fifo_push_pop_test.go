package edge

import "testing"
import "sync"

func TestFIFO_PushPop_FIFOOrder(t *testing.T) {
    fq := &FIFO{MaxCapacity: 0, Backpressure: BackpressureBlock}
    for i := 1; i <= 3; i++ {
        if err := fq.Push(Event{Payload: i}); err != nil { t.Fatalf("push err: %v", err) }
    }
    // Expect 1,2,3 in order
    for i := 1; i <= 3; i++ {
        ev, ok := fq.Pop()
        if !ok { t.Fatalf("pop %d: empty", i) }
        if ev.Payload.(int) != i { t.Fatalf("want %d got %v", i, ev.Payload) }
    }
    if _, ok := fq.Pop(); ok { t.Fatalf("expected empty after pops") }
}

func TestFIFO_Backpressure_DropOldest(t *testing.T) {
    fq := &FIFO{MaxCapacity: 2, Backpressure: BackpressureDrop}
    _ = fq.Push(Event{Payload: 1})
    _ = fq.Push(Event{Payload: 2})
    _ = fq.Push(Event{Payload: 3}) // drop 1
    ev, ok := fq.Pop()
    if !ok || ev.Payload.(int) != 2 { t.Fatalf("expect first=2 got=%v ok=%v", ev.Payload, ok) }
    ev, ok = fq.Pop()
    if !ok || ev.Payload.(int) != 3 { t.Fatalf("expect second=3 got=%v ok=%v", ev.Payload, ok) }
}

func TestFIFO_Backpressure_BlockReturnsErr(t *testing.T) {
    fq := &FIFO{MaxCapacity: 1, Backpressure: BackpressureBlock}
    if err := fq.Push(Event{Payload: 1}); err != nil { t.Fatalf("push err: %v", err) }
    if err := fq.Push(Event{Payload: 2}); err == nil { t.Fatalf("expected ErrFull on block policy") }
}

func TestFIFO_ConcurrentPushes_CountMatches(t *testing.T) {
    const goroutines = 8
    const perG = 100
    fq := &FIFO{MaxCapacity: 0, Backpressure: BackpressureBlock}
    var wg sync.WaitGroup
    wg.Add(goroutines)
    for g := 0; g < goroutines; g++ {
        go func(id int) {
            defer wg.Done()
            for i := 0; i < perG; i++ {
                _ = fq.Push(Event{Payload: i})
            }
        }(g)
    }
    wg.Wait()
    if got := fq.Count(); got != goroutines*perG {
        t.Fatalf("concurrent pushes: want %d got %d", goroutines*perG, got)
    }
}
