package amitime

import (
	"testing"
	stdtime "time"
)

func testNow_CloseToStdlib(t *testing.T) {
	before := stdtime.Now()
	n := Now()
	after := stdtime.Now()
	if n.t.Before(before) || n.t.After(after.Add(2*stdtime.Second)) {
		t.Fatalf("Now() outside expected bounds: before=%v now=%v after=%v", before, n.t, after)
	}
}

func testSleep_Short(t *testing.T) {
	start := stdtime.Now()
	Sleep(20 * stdtime.Millisecond)
	elapsed := stdtime.Since(start)
	if elapsed < 15*stdtime.Millisecond {
		t.Fatalf("sleep too short: %v", elapsed)
	}
}

func testAdd_And_Delta(t *testing.T) {
	t0 := FromUnix(0, 0)
	t1 := Add(t0, stdtime.Second)
	if d := Delta(t0, t1); d != stdtime.Second {
		t.Fatalf("Delta mismatch: got %v want 1s", d)
	}
	if sec := t1.Unix(); sec != 1 {
		t.Fatalf("Unix mismatch: got %d want 1", sec)
	}
}

func testUnixNano(t *testing.T) {
	t0 := FromUnix(1, 500)
	if n := t0.UnixNano(); n <= 1e9 {
		t.Fatalf("expected > 1e9 ns, got %d", n)
	}
}
