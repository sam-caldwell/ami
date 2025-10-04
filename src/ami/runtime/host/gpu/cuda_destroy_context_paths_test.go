package gpu

import "testing"

func TestCudaDestroyContext_Paths(t *testing.T) {
    if err := CudaDestroyContext(Context{}); err != ErrInvalidHandle { t.Fatalf("CudaDestroyContext zero: %v", err) }
    if err := CudaDestroyContext(Context{backend: "cuda", valid: true}); err != ErrUnavailable { t.Fatalf("CudaDestroyContext valid: %v", err) }
}

