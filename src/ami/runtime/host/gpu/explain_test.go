package gpu

import "testing"

func TestExplain_FilePair(t *testing.T) {
    s := Explain("cuda", "alloc", ErrInvalidHandle)
    if s == "" { t.Fatalf("empty explain") }
}

