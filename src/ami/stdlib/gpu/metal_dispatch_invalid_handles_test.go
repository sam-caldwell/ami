package gpu

import "testing"

func TestMetal_Dispatch_InvalidHandles(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
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

