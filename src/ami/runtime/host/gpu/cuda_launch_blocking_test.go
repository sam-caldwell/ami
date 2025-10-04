package gpu

import "testing"

func TestCudaLaunchBlocking_FilePair(t *testing.T) {
    _ = CudaLaunchBlocking(Context{}, Kernel{}, [3]uint32{1,1,1}, [3]uint32{1,1,1}, 0)
}

