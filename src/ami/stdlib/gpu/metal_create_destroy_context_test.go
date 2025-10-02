//go:build darwin

package gpu

import "testing"

func TestMetal_Create_Destroy_Context(t *testing.T) {
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no Metal devices found") }
    ctx, err := MetalCreateContext(devs[0])
    if err != nil { t.Fatalf("MetalCreateContext failed: %v", err) }
    if !ctx.valid || ctx.backend != "metal" { t.Fatalf("context not initialized as expected: %+v", ctx) }
    if err := MetalDestroyContext(ctx); err != nil { t.Fatalf("MetalDestroyContext returned error: %v", err) }
    if err := (&ctx).Release(); err != nil { t.Fatalf("context Release() failed: %v", err) }
    if err := (&ctx).Release(); err == nil { t.Fatalf("double Release() should error") }
}

