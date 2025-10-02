package gpu

import "testing"

// Validates Owned<T> semantics for OpenCL-like handles using stubs.

func TestOpenCL_Context_Release_DeterministicZeroizeAndDoubleFree(t *testing.T) {
    c := &Context{backend: "opencl", valid: true, ctxId: 9}
    if err := c.Release(); err != nil {
        t.Fatalf("first context Release() should succeed; got %v", err)
    }
    if c.valid || c.backend != "" || c.ctxId != 0 {
        t.Fatalf("context not zeroized after release: %+v", c)
    }
    if err := c.Release(); err != ErrInvalidHandle {
        t.Fatalf("second context Release() should be ErrInvalidHandle; got %v", err)
    }
}

func TestOpenCL_Buffer_Release_DeterministicZeroizeAndDoubleFree(t *testing.T) {
    b := &Buffer{backend: "opencl", n: 32, valid: true, bufId: 5}
    if err := b.Release(); err != nil {
        t.Fatalf("first buffer Release() should succeed; got %v", err)
    }
    if b.valid || b.backend != "" || b.n != 0 || b.bufId != 0 {
        t.Fatalf("buffer not zeroized after release: %+v", b)
    }
    if err := b.Release(); err != ErrInvalidHandle {
        t.Fatalf("second buffer Release() should be ErrInvalidHandle; got %v", err)
    }
}

func TestOpenCL_ProgramAndKernel_Release(t *testing.T) {
    // Program
    p := &Program{valid: true}
    if err := p.Release(); err != nil { t.Fatalf("first program Release: %v", err) }
    if err := p.Release(); err != ErrInvalidHandle { t.Fatalf("double-free program: %v", err) }

    // Kernel
    k := &Kernel{valid: true}
    if err := k.Release(); err != nil { t.Fatalf("first kernel Release: %v", err) }
    if err := k.Release(); err != ErrInvalidHandle { t.Fatalf("double-free kernel: %v", err) }
}

