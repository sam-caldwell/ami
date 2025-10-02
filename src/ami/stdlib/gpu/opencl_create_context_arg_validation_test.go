package gpu

import "testing"

func TestOpenCL_CreateContext_ArgValidation(t *testing.T) {
    if _, err := OpenCLCreateContext(Platform{}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLCreateContext empty platform: %v", err)
    }
}

