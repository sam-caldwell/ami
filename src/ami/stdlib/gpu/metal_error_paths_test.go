package gpu

import "testing"

func TestMetal_Alloc_InvalidArg(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    if _, err := MetalAlloc(0); err != ErrInvalidHandle {
        t.Fatalf("MetalAlloc(0): want ErrInvalidHandle, got %v", err)
    }
}

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

func TestMetal_Pipeline_InvalidHandles(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    // invalid library
    if _, err := MetalCreatePipeline(Library{}, "k"); err != ErrInvalidHandle {
        t.Fatalf("MetalCreatePipeline invalid lib: %v", err)
    }
    if _, err := MetalCreatePipeline(Library{valid: true, libId: 0}, "k"); err != ErrInvalidHandle {
        t.Fatalf("MetalCreatePipeline zero id: %v", err)
    }
    // Destroy context invalid
    if err := MetalDestroyContext(Context{}); err != ErrInvalidHandle {
        t.Fatalf("MetalDestroyContext invalid: %v", err)
    }
}

func TestMetal_Dispatch_InvalidHandles(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    // create minimal valid context and library/pipeline, then corrupt handles
    ctx, err := MetalCreateContext(devs[0])
    if err != nil { t.Skip("ctx fail") }
    lib, err := MetalCompileLibrary(mslNoop)
    if err != nil { t.Fatalf("MetalCompileLibrary: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "k0")
    if err != nil { t.Fatalf("MetalCreatePipeline: %v", err) }
    if err := MetalDispatch(Context{}, pipe, [3]uint32{1,1,1}, [3]uint32{1,1,1}); err != ErrInvalidHandle {
        t.Fatalf("MetalDispatch invalid ctx: %v", err)
    }
    if err := MetalDispatch(ctx, Pipeline{}, [3]uint32{1,1,1}, [3]uint32{1,1,1}); err != ErrInvalidHandle {
        t.Fatalf("MetalDispatch invalid pipe: %v", err)
    }
    _ = ctx.Release(); _ = pipe.Release(); _ = lib.Release()
}

