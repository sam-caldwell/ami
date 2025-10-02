package gpu

import (
    "testing"
    "time"
)

func TestBlocking_Blocks(t *testing.T) {
    start := time.Now()
    err := Blocking(func() error { time.Sleep(30 * time.Millisecond); return nil })
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if time.Since(start) < 25*time.Millisecond { t.Fatalf("Blocking did not block as expected") }
}

