//go:build darwin

package gpu

import "testing"

const mslAddScalarI16 = `#include <metal_stdlib>
using namespace metal;

kernel void add_scalar_i16(device short* buf [[buffer(0)]],
                           constant short& c [[buffer(1)]],
                           uint tid [[thread_position_in_grid]]) {
    buf[tid] = buf[tid] + c;
}`

func TestMetal_Dispatch_WithScalar_Int16(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslAddScalarI16); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "add_scalar_i16"); if err != nil { t.Fatalf("pipe: %v", err) }
    const n = 16
    // buffer of shorts (int16)
    buf, err := MetalAlloc(n * 2)
    if err != nil { t.Fatalf("alloc: %v", err) }
    host := make([]byte, n*2)
    for i := 0; i < n; i++ {
        host[i*2] = byte(i+1)
        host[i*2+1] = 0
    }
    if err := MetalCopyToDevice(buf, host); err != nil { t.Fatalf("copy to device: %v", err) }
    add := int16(3)
    grid := [3]uint32{n,1,1}
    tpg := [3]uint32{8,1,1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, buf, add); err != nil {
        t.Fatalf("dispatch: %v", err)
    }
}

