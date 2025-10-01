package amitime

import (
    stdtime "time"
    "sync"
)

// Ticker periodically invokes registered handlers at a fixed interval.
type Ticker struct {
    d        Duration
    mu       sync.Mutex
    started  bool
    stopCh   chan struct{}
    handlers []func()
}

// NewTicker constructs a Ticker with period d.
func NewTicker(d Duration) *Ticker { return &Ticker{d: d} }

// Register adds a handler to be called on each tick.
func (t *Ticker) Register(fn func()) {
    if t == nil || fn == nil { return }
    t.mu.Lock(); defer t.mu.Unlock()
    t.handlers = append(t.handlers, fn)
}

// Start begins periodic ticking. It is idempotent.
func (t *Ticker) Start() {
    if t == nil || t.d <= 0 { return }
    t.mu.Lock()
    if t.started { t.mu.Unlock(); return }
    t.started = true
    t.stopCh = make(chan struct{})
    d := t.d
    t.mu.Unlock()
    go func(){
        tick := stdtime.NewTicker(d)
        defer tick.Stop()
        for {
            select {
            case <-tick.C:
                t.mu.Lock()
                fns := append([]func(){}, t.handlers...)
                t.mu.Unlock()
                for _, f := range fns { safe(f) }
            case <-t.stopCh:
                return
            }
        }
    }()
}

// Stop halts ticking. It is idempotent.
func (t *Ticker) Stop() {
    if t == nil { return }
    t.mu.Lock()
    if !t.started { t.mu.Unlock(); return }
    close(t.stopCh)
    t.started = false
    t.mu.Unlock()
}

func safe(f func()) { defer func(){ _ = recover() }(); if f != nil { f() } }
