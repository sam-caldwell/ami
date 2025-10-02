package gpu

import "testing"

func TestOpenCL_Alloc_ArgValidation(t *testing.T) {
    if _, err := OpenCLAlloc(0); err != ErrInvalidHandle {
        t.Fatalf("OpenCLAlloc(0): want ErrInvalidHandle, got %v", err)
    }
    if _, err := OpenCLAlloc(-1); err != ErrInvalidHandle {
        t.Fatalf("OpenCLAlloc(-1): want ErrInvalidHandle, got %v", err)
    }
}

func TestOpenCL_CreateContext_ArgValidation(t *testing.T) {
    if _, err := OpenCLCreateContext(Platform{}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLCreateContext empty platform: %v", err)
    }
}

func TestOpenCL_ProgramKernel_ArgValidation(t *testing.T) {
    if _, err := OpenCLBuildProgram(""); err != ErrInvalidHandle {
        t.Fatalf("OpenCLBuildProgram empty: %v", err)
    }
    if _, err := OpenCLGetKernel(Program{}, "k"); err != ErrInvalidHandle {
        t.Fatalf("OpenCLGetKernel invalid program: %v", err)
    }
    if _, err := OpenCLGetKernel(Program{valid: true}, ""); err != ErrInvalidHandle {
        t.Fatalf("OpenCLGetKernel empty name: %v", err)
    }
}

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

