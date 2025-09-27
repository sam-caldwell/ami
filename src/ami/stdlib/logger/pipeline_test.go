package logger

import (
    "bytes"
    "testing"
    "time"
)

type fakeSink struct{ buf *bytes.Buffer }

func (f *fakeSink) Start() error { if f.buf == nil { f.buf = &bytes.Buffer{} }; return nil }
func (f *fakeSink) Write(line []byte) error { _, err := f.buf.Write(line); return err }
func (f *fakeSink) Close() error { return nil }

func TestPipeline_BatchFlushOnSizeAndClose(t *testing.T) {
    fs := &fakeSink{}
    p := NewPipeline(Config{Capacity: 10, BatchMax: 2, FlushInterval: 0, Policy: Block, Sink: fs})
    if err := p.Start(); err != nil { t.Fatalf("start: %v", err) }
    if err := p.Enqueue([]byte("a\n")); err != nil { t.Fatalf("enqueue a: %v", err) }
    if err := p.Enqueue([]byte("b\n")); err != nil { t.Fatalf("enqueue b: %v", err) }
    if err := p.Enqueue([]byte("c\n")); err != nil { t.Fatalf("enqueue c: %v", err) }
    p.Close()
    got := fs.buf.String()
    if got != "a\nb\nc\n" {
        t.Fatalf("unexpected writes: %q", got)
    }
}

func TestPipeline_DropNewestWhenQueueFull(t *testing.T) {
    fs := &fakeSink{}
    p := NewPipeline(Config{Capacity: 0, BatchMax: 10, FlushInterval: time.Second, Policy: DropNewest, Sink: fs})
    // Do not start to force the queue to be full/unavailable.
    if err := p.Enqueue([]byte("x\n")); err == nil {
        t.Fatalf("expected drop when not started and capacity=0")
    }
}

