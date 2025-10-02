package gpu

import "testing"

func TestCudaCreateContext_FilePair(t *testing.T) {
    _, _ = CudaCreateContext(Device{Backend: "cuda", ID: 0})
}

