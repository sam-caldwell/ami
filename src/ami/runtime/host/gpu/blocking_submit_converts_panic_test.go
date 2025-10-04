package gpu

import "testing"

func TestBlockingSubmit_ConvertsPanic(t *testing.T) {
    err := BlockingSubmit(func(done chan<- error) { panic("kaboom") })
    if err == nil { t.Fatalf("expected error from panic; got nil") }
}

