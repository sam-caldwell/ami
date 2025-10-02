package gpu

import "testing"

func TestCudaLaunchBlocking_ReturnsUnavailable(t *testing.T) {
    // Provide valid-looking handles so validation passes and stub returns ErrUnavailable.
    err := CudaLaunchBlocking(Context{backend: "cuda", valid: true}, Kernel{valid: true}, [3]uint32{1, 1, 1}, [3]uint32{1, 1, 1}, 0)
    if err != ErrUnavailable {
        t.Fatalf("CudaLaunchBlocking: want ErrUnavailable, got %v", err)
    }
}

func TestOpenCLLaunchBlocking_ReturnsUnavailable(t *testing.T) {
    // Provide valid-looking handles so validation passes and stub returns ErrUnavailable.
    err := OpenCLLaunchBlocking(Context{backend: "opencl", valid: true}, Kernel{valid: true}, [3]uint64{1, 1, 1}, [3]uint64{1, 1, 1})
    if err != ErrUnavailable {
        t.Fatalf("OpenCLLaunchBlocking: want ErrUnavailable, got %v", err)
    }
}
