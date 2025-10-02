package gpu

import "testing"

func TestCuda_Alloc_ArgValidation(t *testing.T) {
    if _, err := CudaAlloc(0); err != ErrInvalidHandle {
        t.Fatalf("CudaAlloc(0): want ErrInvalidHandle, got %v", err)
    }
    if _, err := CudaAlloc(-1); err != ErrInvalidHandle {
        t.Fatalf("CudaAlloc(-1): want ErrInvalidHandle, got %v", err)
    }
}

func TestCuda_Memcpy_ArgValidation(t *testing.T) {
    // Invalid buffer
    if err := CudaMemcpyHtoD(Buffer{}, []byte{1}); err != ErrInvalidHandle {
        t.Fatalf("CudaMemcpyHtoD invalid dst: %v", err)
    }
    if err := CudaMemcpyDtoH([]byte{0}, Buffer{}); err != ErrInvalidHandle {
        t.Fatalf("CudaMemcpyDtoH invalid src: %v", err)
    }
    // Wrong backend
    if err := CudaMemcpyHtoD(Buffer{backend: "opencl", valid: true}, []byte{1}); err != ErrInvalidHandle {
        t.Fatalf("CudaMemcpyHtoD wrong backend: %v", err)
    }
    if err := CudaMemcpyDtoH([]byte{0}, Buffer{backend: "opencl", valid: true}); err != ErrInvalidHandle {
        t.Fatalf("CudaMemcpyDtoH wrong backend: %v", err)
    }
    // Zero length
    if err := CudaMemcpyHtoD(Buffer{backend: "cuda", valid: true}, nil); err != ErrInvalidHandle {
        t.Fatalf("CudaMemcpyHtoD zero length: %v", err)
    }
    if err := CudaMemcpyDtoH(nil, Buffer{backend: "cuda", valid: true}); err != ErrInvalidHandle {
        t.Fatalf("CudaMemcpyDtoH zero length: %v", err)
    }
}

func TestCuda_ModuleKernel_ArgValidation(t *testing.T) {
    if _, err := CudaLoadModule(""); err != ErrInvalidHandle {
        t.Fatalf("CudaLoadModule empty: %v", err)
    }
    if _, err := CudaGetKernel(Module{}, "k"); err != ErrInvalidHandle {
        t.Fatalf("CudaGetKernel invalid mod: %v", err)
    }
    if _, err := CudaGetKernel(Module{valid: true}, ""); err != ErrInvalidHandle {
        t.Fatalf("CudaGetKernel empty name: %v", err)
    }
}

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

