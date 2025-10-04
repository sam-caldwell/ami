package gpu

import "testing"

func TestCudaGetKernel_FilePair(t *testing.T) { _, _ = CudaGetKernel(Module{}, "k") }

