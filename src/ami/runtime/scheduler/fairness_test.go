package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// Property-ish test: FAIR scheduling should distribute work across sources.
func testPool_FAIR_Distribution(t *testing.T) {
	p, err := New(Config{Workers: 3, Policy: FAIR})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()
	var a, b, c int64
	// Submit tasks labeled by source; equal counts
	const N = 60
	for i := 0; i < N; i++ {
		_ = p.Submit(Task{Source: "a", Do: func(ctx context.Context) { atomic.AddInt64(&a, 1) }})
		_ = p.Submit(Task{Source: "b", Do: func(ctx context.Context) { atomic.AddInt64(&b, 1) }})
		_ = p.Submit(Task{Source: "c", Do: func(ctx context.Context) { atomic.AddInt64(&c, 1) }})
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&a) == N && atomic.LoadInt64(&b) == N && atomic.LoadInt64(&c) == N {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("fair distribution not reached: a=%d b=%d c=%d", a, b, c)
}

// Sanity check: WORKSTEAL completes all tasks under load.
func testPool_Worksteal_Completes(t *testing.T) {
	p, err := New(Config{Workers: 4, Policy: WORKSTEAL})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()
	var done int64
	const N = 500
	for i := 0; i < N; i++ {
		_ = p.Submit(Task{Source: "x", Do: func(ctx context.Context) { atomic.AddInt64(&done, 1) }})
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&done) == N {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("worksteal did not complete: done=%d", done)
}
