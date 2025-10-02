package gpu

import (
    "testing"
    "time"
)

func TestBlockingSubmit_Blocks_And_Propagates(t *testing.T) {
    start := time.Now()
    err := BlockingSubmit(func(done chan<- error) {
        go func(){ time.Sleep(20*time.Millisecond); done <- nil }()
    })
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if time.Since(start) < 15*time.Millisecond { t.Fatalf("BlockingSubmit did not block as expected") }
}

