package gpu

import (
    "errors"
    "testing"
    "time"
)

func TestBlocking_PropagatesError(t *testing.T) {
    want := errors.New("x")
    got := Blocking(func() error { return want })
    if !errors.Is(got, want) { t.Fatalf("Blocking() error mismatch: %v", got) }
}

func TestBlocking_ConvertsPanic(t *testing.T) {
    err := Blocking(func() error { panic("boom") })
    if err == nil || err.Error() == "" { t.Fatalf("expected panic converted to error; got %v", err) }
}

func TestBlocking_Blocks(t *testing.T) {
    start := time.Now()
    err := Blocking(func() error { time.Sleep(30 * time.Millisecond); return nil })
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if time.Since(start) < 25*time.Millisecond { t.Fatalf("Blocking did not block as expected") }
}

func TestBlockingSubmit_Blocks_And_Propagates(t *testing.T) {
    start := time.Now()
    err := BlockingSubmit(func(done chan<- error) {
        go func(){ time.Sleep(20*time.Millisecond); done <- nil }()
    })
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if time.Since(start) < 15*time.Millisecond { t.Fatalf("BlockingSubmit did not block as expected") }
}

func TestBlockingSubmit_ConvertsPanic(t *testing.T) {
    err := BlockingSubmit(func(done chan<- error) { panic("kaboom") })
    if err == nil { t.Fatalf("expected error from panic; got nil") }
}

