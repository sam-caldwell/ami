package gpu

import "testing"

func TestCuda_ErrorPaths_BackendMismatch(t *testing.T) {
    if err := CudaDestroyContext(Context{backend: "opencl", valid: true}); err != ErrInvalidHandle {
        t.Fatalf("CudaDestroyContext wrong backend: want ErrInvalidHandle, got %v", err)
    }
    if err := CudaFree(Buffer{backend: "opencl", valid: true}); err != ErrInvalidHandle {
        t.Fatalf("CudaFree wrong backend: want ErrInvalidHandle, got %v", err)
    }
}

