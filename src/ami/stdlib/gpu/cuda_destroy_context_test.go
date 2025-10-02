package gpu

import "testing"

func TestCudaDestroyContext_FilePair(t *testing.T) {
    _ = CudaDestroyContext(Context{})
}

