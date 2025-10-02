package amitime

import (
	"testing"
	stdtime "time"
)

// Ensure Start is idempotent and does not double-fire handlers.
func testTicker_Start_Idempotent(t *testing.T) {
	tk := NewTicker(5 * stdtime.Millisecond)
	defer tk.Stop()
	var count int
	tk.Register(func() { count++ })
	tk.Start()
	tk.Start() // second call should be a no-op
	stdtime.Sleep(30 * stdtime.Millisecond)
	if count == 0 {
		t.Fatalf("expected ticks, got %d", count)
	}
}

// Ensure Stop is idempotent and safe to call multiple times.
func testTicker_Stop_Idempotent(t *testing.T) {
	tk := NewTicker(5 * stdtime.Millisecond)
	var count int
	tk.Register(func() { count++ })
	tk.Start()
	stdtime.Sleep(15 * stdtime.Millisecond)
	tk.Stop()
	tk.Stop() // second stop should not panic
	c1 := count
	stdtime.Sleep(20 * stdtime.Millisecond)
	if count != c1 {
		t.Fatalf("ticks after Stop: before=%d after=%d", c1, count)
	}
}

// Zero or negative durations should not start ticking.
func testTicker_ZeroOrNegative_NoStart(t *testing.T) {
	tk0 := NewTicker(0)
	tkN := NewTicker(-1)
	var c0, cN int
	tk0.Register(func() { c0++ })
	tkN.Register(func() { cN++ })
	tk0.Start()
	tkN.Start()
	defer tk0.Stop()
	defer tkN.Stop()
	stdtime.Sleep(20 * stdtime.Millisecond)
	if c0 != 0 || cN != 0 {
		t.Fatalf("unexpected ticks: c0=%d cN=%d", c0, cN)
	}
}

// Handlers that panic must be recovered and should not crash the ticker goroutine.
func testTicker_HandlerPanicRecovered(t *testing.T) {
	tk := NewTicker(5 * stdtime.Millisecond)
	defer tk.Stop()
	var ok int
	tk.Register(func() { panic("boom") })
	tk.Register(func() { ok++ })
	tk.Start()
	deadline := stdtime.Now().Add(60 * stdtime.Millisecond)
	for stdtime.Now().Before(deadline) {
		if ok > 0 {
			return
		}
		stdtime.Sleep(2 * stdtime.Millisecond)
	}
	t.Fatalf("expected non-panicking handler to run; ok=%d", ok)
}

// Starting after Stop should preserve previously registered handlers.
func testTicker_Restart_PreservesHandlers(t *testing.T) {
	tk := NewTicker(5 * stdtime.Millisecond)
	defer tk.Stop()
	var count int
	tk.Register(func() { count++ })
	tk.Start()
	stdtime.Sleep(12 * stdtime.Millisecond)
	tk.Stop()
	c1 := count
	// restart
	tk.Start()
	stdtime.Sleep(12 * stdtime.Millisecond)
	if count <= c1 {
		t.Fatalf("expected additional ticks after restart: before=%d after=%d", c1, count)
	}
}
