//go:build darwin

package gpu

import "testing"

func TestMetal_Dispatch_UnsupportedArgType_Error(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Skip("ctx fail") }
    lib, err := MetalCompileLibrary(mslNoop); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "k0"); if err != nil { t.Fatalf("pipe: %v", err) }
    // Unsupported arg: boolean true (not mapped in toBytes)
    err = MetalDispatch(ctx, pipe, [3]uint32{1,1,1}, [3]uint32{1,1,1}, true)
    if err == nil { t.Fatalf("expected error for unsupported arg type") }
}

