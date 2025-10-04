package gpu

import "testing"

func TestCudaFree_Paths(t *testing.T) {
    if err := CudaFree(Buffer{}); err != ErrInvalidHandle { t.Fatalf("CudaFree zero: %v", err) }
    if err := CudaFree(Buffer{backend: "cuda", valid: true}); err != ErrUnavailable { t.Fatalf("CudaFree valid: %v", err) }
}

