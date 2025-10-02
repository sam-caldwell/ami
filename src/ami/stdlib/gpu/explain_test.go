package gpu

import "testing"

func TestExplain_DeterministicMessages(t *testing.T) {
    if got := CudaExplain("alloc", ErrInvalidHandle); got != "gpu/cuda alloc: invalid handle" {
        t.Fatalf("cuda explain invalid: %q", got)
    }
    if got := CudaExplain("launch", ErrUnavailable); got != "gpu/cuda launch: backend unavailable" {
        t.Fatalf("cuda explain unavailable: %q", got)
    }
    if got := OpenCLExplain("build", ErrUnimplemented); got != "gpu/opencl build: unimplemented" {
        t.Fatalf("opencl explain unimplemented: %q", got)
    }
    if got := MetalExplain("pipeline", nil); got != "gpu/metal pipeline: ok" {
        t.Fatalf("metal explain ok: %q", got)
    }
}

