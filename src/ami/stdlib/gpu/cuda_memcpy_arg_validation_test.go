package gpu

import "testing"

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

