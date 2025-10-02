package gpu

import "testing"

func TestCudaDestroyContext_Paths(t *testing.T) {
    if err := CudaDestroyContext(Context{}); err != ErrInvalidHandle {
        t.Fatalf("CudaDestroyContext zero: %v", err)
    }
    if err := CudaDestroyContext(Context{backend: "cuda", valid: true}); err != ErrUnavailable {
        t.Fatalf("CudaDestroyContext valid: %v", err)
    }
}

func TestCudaFree_Paths(t *testing.T) {
    if err := CudaFree(Buffer{}); err != ErrInvalidHandle {
        t.Fatalf("CudaFree zero: %v", err)
    }
    if err := CudaFree(Buffer{backend: "cuda", valid: true}); err != ErrUnavailable {
        t.Fatalf("CudaFree valid: %v", err)
    }
}

func TestOpenCLFree_Paths(t *testing.T) {
    if err := OpenCLFree(Buffer{}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLFree zero: %v", err)
    }
    if err := OpenCLFree(Buffer{backend: "opencl", valid: true}); err != ErrUnavailable {
        t.Fatalf("OpenCLFree valid: %v", err)
    }
}

