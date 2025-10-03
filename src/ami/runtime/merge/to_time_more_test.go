package merge

import (
    "testing"
    "time"
)

func Test_toTime(t *testing.T) {
    now := time.Now().UTC().Truncate(time.Second)
    if got, ok := toTime(now); !ok || !got.Equal(now) { t.Fatalf("time pass-through failed") }
    if got, ok := toTime(now.Format(time.RFC3339)); !ok || got.IsZero() { t.Fatalf("rfc3339 parse failed: %v ok=%v", got, ok) }
    if _, ok := toTime("not-a-time"); ok { t.Fatalf("expected false for invalid time string") }
}

