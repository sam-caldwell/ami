package gpu

import "testing"

func TestMetal_Alloc_InvalidArg(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    if _, err := MetalAlloc(0); err != ErrInvalidHandle {
        t.Fatalf("MetalAlloc(0): want ErrInvalidHandle, got %v", err)
    }
}

