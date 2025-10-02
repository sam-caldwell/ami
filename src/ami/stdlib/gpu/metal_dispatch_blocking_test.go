package gpu

import "testing"

func TestMetalDispatchBlocking_FilePair(t *testing.T) {
    _ = MetalDispatchBlocking(Context{}, Pipeline{}, [3]uint32{1,1,1}, [3]uint32{1,1,1})
}

