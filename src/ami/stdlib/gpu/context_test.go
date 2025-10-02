package gpu

import "testing"

func TestContext_FilePair(t *testing.T) {
    var c Context
    if err := c.Release(); err == nil {
        t.Fatalf("expected invalid handle error")
    }
}

