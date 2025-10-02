package gpu

import "testing"

func TestBuffer_FilePair(t *testing.T) {
    var b Buffer
    if err := b.Release(); err == nil { t.Fatalf("expected error") }
}

