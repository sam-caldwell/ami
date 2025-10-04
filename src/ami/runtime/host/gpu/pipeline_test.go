package gpu

import "testing"

func TestPipeline_FilePair(t *testing.T) {
    var p Pipeline
    if err := p.Release(); err == nil { t.Fatalf("expected error") }
}

