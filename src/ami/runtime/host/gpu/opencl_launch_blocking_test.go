package gpu

import "testing"

func TestOpenCLLaunchBlocking_FilePair(t *testing.T) {
    _ = OpenCLLaunchBlocking(Context{}, Kernel{}, [3]uint64{1,1,1}, [3]uint64{1,1,1})
}

