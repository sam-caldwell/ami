package gpu

import "testing"

func TestOpenCL_Alloc_ArgValidation(t *testing.T) {
    if _, err := OpenCLAlloc(0); err != ErrInvalidHandle {
        t.Fatalf("OpenCLAlloc(0): want ErrInvalidHandle, got %v", err)
    }
    if _, err := OpenCLAlloc(-1); err != ErrInvalidHandle {
        t.Fatalf("OpenCLAlloc(-1): want ErrInvalidHandle, got %v", err)
    }
}

