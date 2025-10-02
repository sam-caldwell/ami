package gpu

import "testing"

func TestOpenCLFree_Paths(t *testing.T) {
    if err := OpenCLFree(Buffer{}); err != ErrInvalidHandle { t.Fatalf("OpenCLFree zero: %v", err) }
    if err := OpenCLFree(Buffer{backend: "opencl", valid: true}); err != ErrUnavailable { t.Fatalf("OpenCLFree valid: %v", err) }
}

