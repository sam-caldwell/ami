package driver

import "testing"

func TestCollectEdges_NilFile_ReturnsEmpty(t *testing.T) {
    if out := collectEdges("unit", nil); len(out) != 0 {
        t.Fatalf("expected empty; got %d", len(out))
    }
}

