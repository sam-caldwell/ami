package trigger

import (
	"testing"
	stdtime "time"

	amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
)

func testTimer_EmitsAndStops(t *testing.T) {
	ch, stop := Timer(10 * amitime.Duration(stdtime.Millisecond))
	defer stop()
	// Expect at least 2 ticks within 100ms
	count := 0
	deadline := stdtime.After(150 * stdtime.Millisecond)
	for {
		select {
		case <-ch:
			count++
			if count >= 2 {
				return
			}
		case <-deadline:
			t.Fatalf("expected >=2 timer ticks, got %d", count)
		}
	}
}

func testSchedule_EmitsOnce(t *testing.T) {
	at := amitime.Add(amitime.Now(), 30*amitime.Duration(stdtime.Millisecond))
	ch, _ := Schedule(at)
	select {
	case e := <-ch:
		if e.Value.UnixNano() == 0 {
			t.Fatalf("expected schedule event time")
		}
	case <-stdtime.After(200 * stdtime.Millisecond):
		t.Fatalf("timeout waiting for schedule event")
	}
	// Channel should close after single emission
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatalf("expected closed channel after single emission")
		}
	case <-stdtime.After(50 * stdtime.Millisecond):
		t.Fatalf("channel did not close after single emission")
	}
}
