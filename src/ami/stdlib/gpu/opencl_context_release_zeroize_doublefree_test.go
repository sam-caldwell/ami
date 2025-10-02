package gpu

import "testing"

func TestOpenCL_Context_Release_DeterministicZeroizeAndDoubleFree(t *testing.T) {
    c := &Context{backend: "opencl", valid: true, ctxId: 9}
    if err := c.Release(); err != nil { t.Fatalf("first context Release() should succeed; got %v", err) }
    if c.valid || c.backend != "" || c.ctxId != 0 { t.Fatalf("context not zeroized after release: %+v", c) }
    if err := c.Release(); err != ErrInvalidHandle { t.Fatalf("second context Release() should be ErrInvalidHandle; got %v", err) }
}

