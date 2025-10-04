package gpu

import "testing"

func TestOpenCL_Launch_ArgValidation(t *testing.T) {
    if err := OpenCLLaunchKernel(Context{}, Kernel{}, [3]uint64{1, 1, 1}, [3]uint64{1, 1, 1}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLLaunchKernel invalid ctx/kernel: %v", err)
    }
    if err := OpenCLLaunchKernel(Context{backend: "cuda", valid: true}, Kernel{valid: true}, [3]uint64{1, 1, 1}, [3]uint64{1, 1, 1}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLLaunchKernel wrong backend: %v", err)
    }
    if err := OpenCLLaunchKernel(Context{backend: "opencl", valid: true}, Kernel{valid: true}, [3]uint64{0, 1, 1}, [3]uint64{1, 1, 1}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLLaunchKernel zero global: %v", err)
    }
    if err := OpenCLLaunchKernel(Context{backend: "opencl", valid: true}, Kernel{valid: true}, [3]uint64{1, 1, 1}, [3]uint64{0, 1, 1}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLLaunchKernel zero local: %v", err)
    }
}

