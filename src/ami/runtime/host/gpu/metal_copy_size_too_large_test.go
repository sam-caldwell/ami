package gpu

import "testing"

func TestMetal_Copy_SizeTooLarge(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Skip("ctx fail") }
    defer func(){ _ = ctx.Release() }()
    buf, err := MetalAlloc(4)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }
    // to-device too large
    if err := MetalCopyToDevice(buf, []byte{1,2,3,4,5}); err != ErrInvalidHandle {
        t.Fatalf("MetalCopyToDevice oversize: %v", err)
    }
    // from-device too large
    dst := make([]byte, 5)
    if err := MetalCopyFromDevice(dst, buf); err != ErrInvalidHandle {
        t.Fatalf("MetalCopyFromDevice oversize: %v", err)
    }
    // invalid backend
    wrong := Buffer{backend: "opencl", valid: true, bufId: 1, n: 4}
    if err := MetalCopyToDevice(wrong, []byte{1}); err != ErrInvalidHandle {
        t.Fatalf("MetalCopyToDevice wrong backend: %v", err)
    }
}

