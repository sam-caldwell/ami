package gpu

import "testing"

func TestOpenCLLaunchBlocking_ReturnsUnavailable(t *testing.T) {
    // Provide valid-looking handles so validation passes and stub returns ErrUnavailable.
    err := OpenCLLaunchBlocking(Context{backend: "opencl", valid: true}, Kernel{valid: true}, [3]uint64{1, 1, 1}, [3]uint64{1, 1, 1})
    if err != ErrUnavailable { t.Fatalf("OpenCLLaunchBlocking: want ErrUnavailable, got %v", err) }
}

