package gpu

import "testing"

func TestCudaLaunchKernel_FilePair(t *testing.T) { _ = CudaLaunchKernel(Context{}, Kernel{}, [3]uint32{1,1,1}, [3]uint32{1,1,1}, 0) }

