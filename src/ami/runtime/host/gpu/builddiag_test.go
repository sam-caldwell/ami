package gpu

import "testing"

func TestBuildDiag_FilePair(t *testing.T) {
    d := BuildDiag("cuda", "alloc", ErrUnavailable)
    if d.Code == "" { t.Fatalf("empty code") }
}

