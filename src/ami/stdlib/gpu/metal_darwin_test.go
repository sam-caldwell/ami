//go:build darwin

package gpu

import "testing"

func TestMetal_Available_And_Devices(t *testing.T) {
    if !MetalAvailable() {
        t.Fatalf("MetalAvailable() should be true on darwin host with Metal GPU")
    }
    devs := MetalDevices()
    if len(devs) == 0 {
        t.Fatalf("MetalDevices() returned empty list")
    }
    for i, d := range devs {
        if d.Backend != "metal" { t.Fatalf("device[%d].Backend=%q want metal", i, d.Backend) }
        if d.ID < 0 { t.Fatalf("device[%d].ID < 0", i) }
        if d.Name == "" { t.Fatalf("device[%d] has empty Name", i) }
    }
}

func TestMetal_Create_Destroy_Context(t *testing.T) {
    devs := MetalDevices()
    if len(devs) == 0 { t.Skip("no Metal devices found") }
    ctx, err := MetalCreateContext(devs[0])
    if err != nil { t.Fatalf("MetalCreateContext failed: %v", err) }
    if !ctx.valid || ctx.backend != "metal" {
        t.Fatalf("context not initialized as expected: %+v", ctx)
    }
    // Destroy via function (no mutation in copy); should succeed
    if err := MetalDestroyContext(ctx); err != nil {
        t.Fatalf("MetalDestroyContext returned error: %v", err)
    }
    // Release via method (mutates); should succeed then fail on double release
    if err := (&ctx).Release(); err != nil {
        t.Fatalf("context Release() failed: %v", err)
    }
    if err := (&ctx).Release(); err == nil {
        t.Fatalf("double Release() should error")
    }
}

