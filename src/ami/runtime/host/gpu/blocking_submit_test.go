package gpu

import "testing"

func TestBlockingSubmit_FilePair(t *testing.T) {
    err := BlockingSubmit(func(done chan<- error) { done <- nil })
    if err != nil { t.Fatalf("BlockingSubmit returned error: %v", err) }
}

