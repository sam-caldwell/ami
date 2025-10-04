package gpu

import "testing"

func TestCuda_Alloc_ArgValidation(t *testing.T) {
    if _, err := CudaAlloc(0); err != ErrInvalidHandle {
        t.Fatalf("CudaAlloc(0): want ErrInvalidHandle, got %v", err)
    }
    if _, err := CudaAlloc(-1); err != ErrInvalidHandle {
        t.Fatalf("CudaAlloc(-1): want ErrInvalidHandle, got %v", err)
    }
}

