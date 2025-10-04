package gpu

import "testing"

func TestCuda_ModuleKernel_ArgValidation(t *testing.T) {
    if _, err := CudaLoadModule(""); err != ErrInvalidHandle {
        t.Fatalf("CudaLoadModule empty: %v", err)
    }
    if _, err := CudaGetKernel(Module{}, "k"); err != ErrInvalidHandle {
        t.Fatalf("CudaGetKernel invalid mod: %v", err)
    }
    if _, err := CudaGetKernel(Module{valid: true}, ""); err != ErrInvalidHandle {
        t.Fatalf("CudaGetKernel empty name: %v", err)
    }
}

