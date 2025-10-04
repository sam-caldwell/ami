package gpu

import (
    "errors"
    "testing"
)

func TestBlocking_PropagatesError(t *testing.T) {
    want := errors.New("x")
    got := Blocking(func() error { return want })
    if !errors.Is(got, want) { t.Fatalf("Blocking() error mismatch: %v", got) }
}

