package gpu

import "testing"

func TestProgram_FilePair(t *testing.T) {
    var p Program
    if err := p.Release(); err == nil { t.Fatalf("expected error") }
}

