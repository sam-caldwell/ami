package edge

import "testing"

func TestFIFOQueue_OrderAndBackpressure(t *testing.T) {
    // dropNewest policy
    q, err := NewFIFO(FIFO{MaxCapacity: 2, Backpressure: "dropNewest"})
    if err != nil { t.Fatalf("new fifo: %v", err) }
    _ = q.Push(1)
    _ = q.Push(2)
    // third push dropped
    _ = q.Push(3)
    if q.Len() != 2 { t.Fatalf("len: %d", q.Len()) }
    a, ok := q.Pop(); if !ok || a.(int) != 1 { t.Fatalf("pop a: %v %v", a, ok) }
    b, ok := q.Pop(); if !ok || b.(int) != 2 { t.Fatalf("pop b: %v %v", b, ok) }
    if _, ok := q.Pop(); ok { t.Fatalf("expected empty") }

    // dropOldest policy makes room by removing 1
    q2, _ := NewFIFO(FIFO{MaxCapacity: 2, Backpressure: "dropOldest"})
    _ = q2.Push(1); _ = q2.Push(2); _ = q2.Push(3)
    if q2.Len() != 2 { t.Fatalf("len2: %d", q2.Len()) }
    a2, _ := q2.Pop(); if a2.(int) != 2 { t.Fatalf("a2: %v", a2) }
    b2, _ := q2.Pop(); if b2.(int) != 3 { t.Fatalf("b2: %v", b2) }

    // block returns ErrFull
    q3, _ := NewFIFO(FIFO{MaxCapacity: 1, Backpressure: "block"})
    _ = q3.Push(10)
    if err := q3.Push(11); err == nil { t.Fatalf("expected ErrFull") }
}

func TestLIFOStack_OrderAndBackpressure(t *testing.T) {
    s, _ := NewLIFO(LIFO{MaxCapacity: 2, Backpressure: "dropNewest"})
    _ = s.Push(1); _ = s.Push(2); _ = s.Push(3) // drop 3
    a, _ := s.Pop(); if a.(int) != 2 { t.Fatalf("a: %v", a) }
    b, _ := s.Pop(); if b.(int) != 1 { t.Fatalf("b: %v", b) }
    if _, ok := s.Pop(); ok { t.Fatalf("expected empty") }

    s2, _ := NewLIFO(LIFO{MaxCapacity: 2, Backpressure: "dropOldest"})
    _ = s2.Push(1); _ = s2.Push(2); _ = s2.Push(3) // drop oldest (1)
    a2, _ := s2.Pop(); if a2.(int) != 3 { t.Fatalf("a2: %v", a2) }
    b2, _ := s2.Pop(); if b2.(int) != 2 { t.Fatalf("b2: %v", b2) }

    s3, _ := NewLIFO(LIFO{MaxCapacity: 1, Backpressure: "block"})
    _ = s3.Push(9)
    if err := s3.Push(10); err == nil { t.Fatalf("expected ErrFull") }
}

func TestPipelineBuffer_Basics(t *testing.T) {
    p := NewPipelineBuffer()
    _ = p.Push("a"); _ = p.Push("b")
    x, _ := p.Pop(); if x.(string) != "a" { t.Fatalf("x: %v", x) }
    y, _ := p.Pop(); if y.(string) != "b" { t.Fatalf("y: %v", y) }
}

// Simple concurrency: single producer and single consumer over FIFO and LIFO to ensure race-free behavior.
func TestEdgeRuntime_SimpleConcurrency(t *testing.T) {
    // FIFO
    q, _ := NewFIFO(FIFO{MaxCapacity: 0})
    const N = 100
    done := make(chan struct{})
    go func(){
        for i:=0;i<N;i++ { _ = q.Push(i) }
        close(done)
    }()
    <-done
    for i:=0;i<N;i++ { v, ok := q.Pop(); if !ok || v.(int)!=i { t.Fatalf("fifo %d got %v ok=%v", i, v, ok) } }
    // LIFO
    s, _ := NewLIFO(LIFO{MaxCapacity: 0})
    done2 := make(chan struct{})
    go func(){
        for i:=0;i<N;i++ { _ = s.Push(i) }
        close(done2)
    }()
    <-done2
    for i:=N-1;i>=0;i-- { v, ok := s.Pop(); if !ok || v.(int)!=i { t.Fatalf("lifo %d got %v ok=%v", i, v, ok) } }
}
