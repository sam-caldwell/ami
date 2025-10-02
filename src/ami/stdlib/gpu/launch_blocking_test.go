package gpu

import "testing"

func TestCudaLaunchBlocking_ReturnsUnavailable(t *testing.T) {
    // Stubbed backend -> returns ErrUnavailable via Blocking wrapper
    err := CudaLaunchBlocking(Context{}, Kernel{}, [3]uint32{1, 1, 1}, [3]uint32{1, 1, 1}, 0)
    if err != ErrUnavailable {
        t.Fatalf("CudaLaunchBlocking: want ErrUnavailable, got %v", err)
    }
}

func TestOpenCLLaunchBlocking_ReturnsUnavailable(t *testing.T) {
    // Stubbed backend -> returns ErrUnavailable via Blocking wrapper
    err := OpenCLLaunchBlocking(Context{}, Kernel{}, [3]uint64{1, 1, 1}, [3]uint64{1, 1, 1})
    if err != ErrUnavailable {
        t.Fatalf("OpenCLLaunchBlocking: want ErrUnavailable, got %v", err)
    }
}

