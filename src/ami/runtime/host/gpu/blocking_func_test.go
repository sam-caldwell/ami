package gpu

import "testing"

func TestBlocking_FilePair(t *testing.T) {
    if err := Blocking(func() error { return nil }); err != nil {
        t.Fatalf("Blocking returned error: %v", err)
    }
}

