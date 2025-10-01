package gpu

import "testing"

func TestGPUAvailabilityAndDiscovery(t *testing.T) {
    if CudaAvailable() {
        t.Fatalf("CudaAvailable() stub should be false in stub build")
    }
    if OpenCLAvailable() {
        t.Fatalf("OpenCLAvailable() stub should be false in stub build")
    }
    if d := CudaDevices(); d != nil && len(d) != 0 {
        t.Fatalf("CudaDevices() expected empty slice; got %v", d)
    }
    if p := OpenCLPlatforms(); p != nil && len(p) != 0 {
        t.Fatalf("OpenCLPlatforms() expected empty slice; got %v", p)
    }
}

func TestGPUStubs_ReturnUnavailable_AndHandleChecks(t *testing.T) {
    // CUDA
    if _, err := CudaCreateContext(Device{}); err != ErrUnavailable {
        t.Fatalf("CudaCreateContext expected ErrUnavailable; got %v", err)
    }
    if err := CudaDestroyContext(Context{}); err != ErrInvalidHandle {
        t.Fatalf("CudaDestroyContext expected ErrInvalidHandle for zero handle; got %v", err)
    }
    if _, err := CudaAlloc(1024); err != ErrUnavailable {
        t.Fatalf("CudaAlloc expected ErrUnavailable; got %v", err)
    }
    if err := CudaFree(Buffer{}); err != ErrInvalidHandle {
        t.Fatalf("CudaFree expected ErrInvalidHandle for zero handle; got %v", err)
    }
    if err := CudaMemcpyHtoD(Buffer{}, []byte("hi")); err != ErrUnavailable {
        t.Fatalf("CudaMemcpyHtoD expected ErrUnavailable; got %v", err)
    }
    if err := CudaMemcpyDtoH(make([]byte, 2), Buffer{}); err != ErrUnavailable {
        t.Fatalf("CudaMemcpyDtoH expected ErrUnavailable; got %v", err)
    }
    if _, err := CudaLoadModule(".ptx"); err != ErrUnavailable {
        t.Fatalf("CudaLoadModule expected ErrUnavailable; got %v", err)
    }
    if _, err := CudaGetKernel(Module{}, "k"); err != ErrUnavailable {
        t.Fatalf("CudaGetKernel expected ErrUnavailable; got %v", err)
    }
    if err := CudaLaunchKernel(Context{}, Kernel{}, [3]uint32{1,1,1}, [3]uint32{1,1,1}, 0); err != ErrUnavailable {
        t.Fatalf("CudaLaunchKernel expected ErrUnavailable; got %v", err)
    }

    // Metal checks are covered in OS-specific tests.

    // OpenCL
    if _, err := OpenCLCreateContext(Platform{}); err != ErrUnavailable {
        t.Fatalf("OpenCLCreateContext expected ErrUnavailable; got %v", err)
    }
    if _, err := OpenCLAlloc(64); err != ErrUnavailable {
        t.Fatalf("OpenCLAlloc expected ErrUnavailable; got %v", err)
    }
    if err := OpenCLFree(Buffer{}); err != ErrInvalidHandle {
        t.Fatalf("OpenCLFree expected ErrInvalidHandle for zero handle; got %v", err)
    }
    if _, err := OpenCLBuildProgram("src"); err != ErrUnavailable {
        t.Fatalf("OpenCLBuildProgram expected ErrUnavailable; got %v", err)
    }
    if _, err := OpenCLGetKernel(Program{}, "k"); err != ErrUnavailable {
        t.Fatalf("OpenCLGetKernel expected ErrUnavailable; got %v", err)
    }
    if err := OpenCLLaunchKernel(Context{}, Kernel{}, [3]uint64{1,1,1}, [3]uint64{1,1,1}); err != ErrUnavailable {
        t.Fatalf("OpenCLLaunchKernel expected ErrUnavailable; got %v", err)
    }
}

func TestOwnedReleaseSemantics_Stubs(t *testing.T) {
    // Context
    c := &Context{backend: "cuda", valid: true}
    if err := c.Release(); err != nil { t.Fatalf("first context Release() should succeed; got %v", err) }
    if err := c.Release(); err == nil {
        t.Fatalf("second context Release() should error")
    }

    // Buffer
    b := &Buffer{backend: "metal", n: 128, valid: true}
    if err := b.Release(); err != nil { t.Fatalf("first buffer Release() should succeed; got %v", err) }
    if err := b.Release(); err == nil { t.Fatalf("second buffer Release() should error") }

    // Module
    m := &Module{valid: true}
    if err := m.Release(); err != nil { t.Fatalf("first module Release() should succeed; got %v", err) }
    if err := m.Release(); err == nil { t.Fatalf("second module Release() should error") }

    // Kernel
    k := &Kernel{valid: true}
    if err := k.Release(); err != nil { t.Fatalf("first kernel Release() should succeed; got %v", err) }
    if err := k.Release(); err == nil { t.Fatalf("second kernel Release() should error") }

    // Library
    l := &Library{valid: true}
    if err := l.Release(); err != nil { t.Fatalf("first library Release() should succeed; got %v", err) }
    if err := l.Release(); err == nil { t.Fatalf("second library Release() should error") }

    // Pipeline
    p := &Pipeline{valid: true}
    if err := p.Release(); err != nil { t.Fatalf("first pipeline Release() should succeed; got %v", err) }
    if err := p.Release(); err == nil { t.Fatalf("second pipeline Release() should error") }

    // Program
    pr := &Program{valid: true}
    if err := pr.Release(); err != nil { t.Fatalf("first program Release() should succeed; got %v", err) }
    if err := pr.Release(); err == nil { t.Fatalf("second program Release() should error") }
}
