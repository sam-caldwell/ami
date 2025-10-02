package scheduler

import "testing"

func TestNewPoolCreateStop(t *testing.T) {
    p, err := New(Config{Workers: 1, Policy: FIFO})
    if err != nil { t.Fatalf("new: %v", err) }
    p.Stop()
}

