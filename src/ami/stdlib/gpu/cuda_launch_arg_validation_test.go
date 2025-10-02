package gpu

import "testing"

func TestCuda_Launch_ArgValidation(t *testing.T) {
    // Invalid ctx/kernel
    if err := CudaLaunchKernel(Context{}, Kernel{}, [3]uint32{1, 1, 1}, [3]uint32{1, 1, 1}, 0); err != ErrInvalidHandle {
        t.Fatalf("CudaLaunchKernel invalid ctx/kernel: %v", err)
    }
    // Wrong backend
    if err := CudaLaunchKernel(Context{backend: "opencl", valid: true}, Kernel{valid: true}, [3]uint32{1, 1, 1}, [3]uint32{1, 1, 1}, 0); err != ErrInvalidHandle {
        t.Fatalf("CudaLaunchKernel wrong backend: %v", err)
    }
    // Zero dims
    if err := CudaLaunchKernel(Context{backend: "cuda", valid: true}, Kernel{valid: true}, [3]uint32{0, 1, 1}, [3]uint32{1, 1, 1}, 0); err != ErrInvalidHandle {
        t.Fatalf("CudaLaunchKernel zero grid: %v", err)
    }
    if err := CudaLaunchKernel(Context{backend: "cuda", valid: true}, Kernel{valid: true}, [3]uint32{1, 1, 1}, [3]uint32{0, 1, 1}, 0); err != ErrInvalidHandle {
        t.Fatalf("CudaLaunchKernel zero block: %v", err)
    }
}

