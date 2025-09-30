package merge

import (
    "testing"
    "time"
)

func TestLatePolicy_Accept_Vs_Drop(t *testing.T) {
    now := time.Now()
    // Drop policy: late event should be dropped when older than lateness
    var p1 Plan
    p1.Watermark = &Watermark{Field:"ts", LatenessMs: 50}
    p1.LatePolicy = "drop"
    op1 := NewOperator(p1)
    _ = op1.Push(E(map[string]any{"ts": now.Add(-time.Second).Format(time.RFC3339)}))
    if _, ok := op1.Pop(); ok { t.Fatalf("expected drop") }
    // Accept policy: late event should be accepted
    var p2 Plan
    p2.Watermark = &Watermark{Field:"ts", LatenessMs: 50}
    p2.LatePolicy = "accept"
    op2 := NewOperator(p2)
    _ = op2.Push(E(map[string]any{"ts": now.Add(-time.Second).Format(time.RFC3339)}))
    if _, ok := op2.Pop(); !ok { t.Fatalf("expected accept of late event") }
}

