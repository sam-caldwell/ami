package merge

import (
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "testing"
    "time"
)

func TestOperator_ExpireStale_DropsBuffers(t *testing.T) {
    p := Plan{}
    p.TimeoutMs = 1
    op := NewOperator(p)
    _ = op.Push(ev.Event{Payload: map[string]any{"x":1}})
    time.Sleep(2*time.Millisecond)
    if n := op.ExpireStale(time.Now()); n == 0 { t.Fatalf("expected drops > 0") }
}

// Note: FlushByWatermark behavior is covered by golden tests in zz_ files.
