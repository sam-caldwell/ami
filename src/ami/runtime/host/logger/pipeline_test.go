package logger

import (
	"bytes"
	"testing"
	"time"
)

type fakeSink struct{ buf *bytes.Buffer }

func (f *fakeSink) Start() error {
	if f.buf == nil {
		f.buf = &bytes.Buffer{}
	}
	return nil
}
func (f *fakeSink) Write(line []byte) error { _, err := f.buf.Write(line); return err }
func (f *fakeSink) Close() error            { return nil }

func testPipeline_BatchFlushOnSizeAndClose(t *testing.T) {
	fs := &fakeSink{}
	p := NewPipeline(Config{Capacity: 10, BatchMax: 2, FlushInterval: 0, Policy: Block, Sink: fs})
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := p.Enqueue([]byte("a\n")); err != nil {
		t.Fatalf("enqueue a: %v", err)
	}
	if err := p.Enqueue([]byte("b\n")); err != nil {
		t.Fatalf("enqueue b: %v", err)
	}
	if err := p.Enqueue([]byte("c\n")); err != nil {
		t.Fatalf("enqueue c: %v", err)
	}
	p.Close()
	got := fs.buf.String()
	if got != "a\nb\nc\n" {
		t.Fatalf("unexpected writes: %q", got)
	}
}

func testPipeline_DropNewestWhenQueueFull(t *testing.T) {
	fs := &fakeSink{}
	p := NewPipeline(Config{Capacity: 0, BatchMax: 10, FlushInterval: time.Second, Policy: DropNewest, Sink: fs})
	// Do not start to force the queue to be full/unavailable.
	if err := p.Enqueue([]byte("x\n")); err == nil {
		t.Fatalf("expected drop when not started and capacity=0")
	}
}

func testPipeline_DropOldest_Policy(t *testing.T) {
	fs := &fakeSink{}
	// capacity=1, not started yet to deterministically fill channel
	p := NewPipeline(Config{Capacity: 1, BatchMax: 10, FlushInterval: 0, Policy: DropOldest, Sink: fs})
	if err := p.Enqueue([]byte("a\n")); err != nil {
		t.Fatalf("enqueue a: %v", err)
	}
	// second enqueue should drop oldest (a) and keep b
	if err := p.Enqueue([]byte("b\n")); err != nil {
		t.Fatalf("enqueue b: %v", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Close to drain and flush
	p.Close()
	got := fs.buf.String()
	if got != "b\n" {
		t.Fatalf("dropOldest unexpected sink contents: %q", got)
	}
}

func testPipeline_FlushInterval(t *testing.T) {
	fs := &fakeSink{}
	p := NewPipeline(Config{Capacity: 4, BatchMax: 10, FlushInterval: 20 * time.Millisecond, Policy: Block, Sink: fs})
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := p.Enqueue([]byte("x\n")); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	p.Close()
	if fs.buf.String() == "" {
		t.Fatalf("expected time-based flush to write data")
	}
}

func testPipeline_Counters(t *testing.T) {
	fs := &fakeSink{}
	p := NewPipeline(Config{Capacity: 1, BatchMax: 1, FlushInterval: 0, Policy: DropNewest, Sink: fs})
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	_ = p.Enqueue([]byte("a\n"))
	// This will likely drop due to capacity=1 and immediate block in run loop; ensure increment
	_ = p.Enqueue([]byte("b\n"))
	p.Close()
	st := p.Stats()
	if st.Enqueued == 0 || st.Batches == 0 || st.Flushes == 0 {
		t.Fatalf("expected counters incremented: %+v", st)
	}
}
