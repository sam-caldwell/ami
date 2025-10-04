package gpu

import "testing"

func TestModule_FilePair(t *testing.T) {
    var m Module
    if err := m.Release(); err == nil { t.Fatalf("expected error") }
}

