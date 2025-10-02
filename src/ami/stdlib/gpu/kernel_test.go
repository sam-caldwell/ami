package gpu

import "testing"

func TestKernel_FilePair(t *testing.T) {
    var k Kernel
    if err := k.Release(); err == nil { t.Fatalf("expected error") }
}

