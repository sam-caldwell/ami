package gpu

import "testing"

func TestCudaLoadModule_FilePair(t *testing.T) { _, _ = CudaLoadModule(".entry x(){}") }

