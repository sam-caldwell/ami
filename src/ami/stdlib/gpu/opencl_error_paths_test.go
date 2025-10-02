package gpu

import "testing"

func TestOpenCL_ErrorPaths_BackendMismatch(t *testing.T) {
    if err := OpenCLFree(Buffer{backend: "cuda", valid: true}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLFree wrong backend: want ErrInvalidHandle, got %v", err)
    }
}

