package gpu

import "testing"

func TestMetal_Pipeline_InvalidHandles(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    if _, err := MetalCreatePipeline(Library{}, "k"); err != ErrInvalidHandle {
        t.Fatalf("MetalCreatePipeline invalid lib: %v", err)
    }
    if _, err := MetalCreatePipeline(Library{valid: true, libId: 0}, "k"); err != ErrInvalidHandle {
        t.Fatalf("MetalCreatePipeline zero id: %v", err)
    }
    if err := MetalDestroyContext(Context{}); err != ErrInvalidHandle {
        t.Fatalf("MetalDestroyContext invalid: %v", err)
    }
}

