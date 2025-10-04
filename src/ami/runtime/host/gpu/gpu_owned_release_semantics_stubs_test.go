package gpu

import "testing"

func TestOwnedReleaseSemantics_Stubs(t *testing.T) {
    // Context
    c := &Context{backend: "cuda", valid: true}
    if err := c.Release(); err != nil { t.Fatalf("first context Release() should succeed; got %v", err) }
    if err := c.Release(); err == nil { t.Fatalf("second context Release() should error") }

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

