package gpu

import "testing"

func TestOpenCLLaunchKernel_FilePair(t *testing.T) {
    _ = OpenCLLaunchKernel(Context{}, Kernel{}, [3]uint64{1,1,1}, [3]uint64{1,1,1})
}

