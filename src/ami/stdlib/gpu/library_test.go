package gpu

import "testing"

func TestLibrary_FilePair(t *testing.T) {
    var l Library
    if err := l.Release(); err == nil { t.Fatalf("expected error") }
}

