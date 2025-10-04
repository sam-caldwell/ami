package amitime

import (
    stdtime "time"
    "testing"
)

func TestTicker_Basic_StartRegisterStop(t *testing.T) {
    tk := NewTicker(10 * stdtime.Millisecond)
    var count int
    tk.Register(func(){ count++ })
    tk.Start()
    defer tk.Stop()
    // wait up to 200ms for several ticks
    deadline := stdtime.Now().Add(200 * stdtime.Millisecond)
    for stdtime.Now().Before(deadline) {
        if count >= 3 { return }
        stdtime.Sleep(5 * stdtime.Millisecond)
    }
    t.Fatalf("expected count>=3, got %d", count)
}

