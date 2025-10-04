package gpu

import "testing"

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

