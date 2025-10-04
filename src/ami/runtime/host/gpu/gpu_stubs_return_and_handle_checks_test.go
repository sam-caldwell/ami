package gpu

import "testing"

func TestGPUStubs_ReturnUnavailable_AndHandleChecks(t *testing.T) {
    // CUDA
    if _, err := CudaCreateContext(Device{}); err != ErrInvalidHandle { t.Fatalf("CudaCreateContext expected ErrInvalidHandle for empty device; got %v", err) }
    if err := CudaDestroyContext(Context{}); err != ErrInvalidHandle { t.Fatalf("CudaDestroyContext expected ErrInvalidHandle for zero handle; got %v", err) }
    if _, err := CudaAlloc(1024); err != ErrUnavailable { t.Fatalf("CudaAlloc expected ErrUnavailable; got %v", err) }
    if err := CudaFree(Buffer{}); err != ErrInvalidHandle { t.Fatalf("CudaFree expected ErrInvalidHandle for zero handle; got %v", err) }
    if err := CudaMemcpyHtoD(Buffer{}, []byte("hi")); err != ErrInvalidHandle { t.Fatalf("CudaMemcpyHtoD expected ErrInvalidHandle for zero buffer; got %v", err) }
    if err := CudaMemcpyDtoH(make([]byte, 2), Buffer{}); err != ErrInvalidHandle { t.Fatalf("CudaMemcpyDtoH expected ErrInvalidHandle for zero buffer; got %v", err) }
    if _, err := CudaLoadModule(".ptx"); err != ErrUnavailable { t.Fatalf("CudaLoadModule expected ErrUnavailable; got %v", err) }
    if _, err := CudaGetKernel(Module{}, "k"); err != ErrInvalidHandle { t.Fatalf("CudaGetKernel expected ErrInvalidHandle for invalid module; got %v", err) }
    if err := CudaLaunchKernel(Context{}, Kernel{}, [3]uint32{1,1,1}, [3]uint32{1,1,1}, 0); err != ErrInvalidHandle { t.Fatalf("CudaLaunchKernel expected ErrInvalidHandle for invalid ctx/kernel; got %v", err) }

    // OpenCL
    if _, err := OpenCLCreateContext(Platform{}); err != ErrInvalidHandle { t.Fatalf("OpenCLCreateContext expected ErrInvalidHandle for empty platform; got %v", err) }
    if _, err := OpenCLAlloc(64); err != ErrUnavailable { t.Fatalf("OpenCLAlloc expected ErrUnavailable; got %v", err) }
    if err := OpenCLFree(Buffer{}); err != ErrInvalidHandle { t.Fatalf("OpenCLFree expected ErrInvalidHandle for zero handle; got %v", err) }
    if _, err := OpenCLBuildProgram("src"); err != ErrUnavailable { t.Fatalf("OpenCLBuildProgram expected ErrUnavailable; got %v", err) }
    if _, err := OpenCLGetKernel(Program{}, "k"); err != ErrInvalidHandle { t.Fatalf("OpenCLGetKernel expected ErrInvalidHandle for invalid program; got %v", err) }
    if err := OpenCLLaunchKernel(Context{}, Kernel{}, [3]uint64{1,1,1}, [3]uint64{1,1,1}); err != ErrInvalidHandle { t.Fatalf("OpenCLLaunchKernel expected ErrInvalidHandle for invalid ctx/kernel; got %v", err) }
}

