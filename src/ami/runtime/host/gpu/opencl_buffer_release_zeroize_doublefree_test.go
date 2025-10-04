package gpu

import "testing"

func TestOpenCL_Buffer_Release_DeterministicZeroizeAndDoubleFree(t *testing.T) {
    b := &Buffer{backend: "opencl", n: 32, valid: true, bufId: 5}
    if err := b.Release(); err != nil { t.Fatalf("first buffer Release() should succeed; got %v", err) }
    if b.valid || b.backend != "" || b.n != 0 || b.bufId != 0 { t.Fatalf("buffer not zeroized after release: %+v", b) }
    if err := b.Release(); err != ErrInvalidHandle { t.Fatalf("second buffer Release() should be ErrInvalidHandle; got %v", err) }
}

