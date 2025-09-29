package scheduler

import (
    "testing"
    "sync/atomic"
    "time"
    "context"
)

func TestPool_FIFO(t *testing.T) {
    p, err := New(Config{Workers: 2, Policy: FIFO, QueueCapacity: 10})
    if err != nil { t.Fatal(err) }
    defer p.Stop()
    var c int64
    for i := 0; i < 10; i++ {
        _ = p.Submit(Task{Do: func(ctx context.Context){ atomic.AddInt64(&c, 1) }})
    }
    deadline := time.Now().Add(1*time.Second)
    for time.Now().Before(deadline) {
        if atomic.LoadInt64(&c) == 10 { return }
        time.Sleep(10*time.Millisecond)
    }
    t.Fatalf("expected 10 tasks, got %d", c)
}
