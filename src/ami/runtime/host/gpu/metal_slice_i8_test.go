//go:build darwin

package gpu

import (
    "testing"
)

const mslIncFromI8Slice = `#include <metal_stdlib>
using namespace metal;

kernel void inc_from_i8_slice(device char* out [[buffer(0)]],
                              constant char* arr [[buffer(1)]],
                              constant uint& n [[buffer(2)]],
                              uint tid [[thread_position_in_grid]]) {
    if (tid < n) {
        out[tid] = arr[tid] + (char)1;
    }
}`

func TestMetal_Dispatch_WithSlice_Int8(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslIncFromI8Slice); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "inc_from_i8_slice"); if err != nil { t.Fatalf("pipe: %v", err) }

    const n = 32
    in := make([]int8, n)
    for i := 0; i < n; i++ { in[i] = int8(i) }
    outBuf, err := MetalAlloc(n)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }

    grid := [3]uint32{n,1,1}
    tpg  := [3]uint32{16,1,1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, outBuf, in, uint32(n)); err != nil {
        t.Fatalf("dispatch: %v", err)
    }
    out := make([]byte, n)
    if err := MetalCopyFromDevice(out, outBuf); err != nil { t.Fatalf("copy from device: %v", err) }
    for i := 0; i < n; i++ {
        want := int8(i+1)
        if int8(out[i]) != want { t.Fatalf("out[%d]=%d want %d", i, int8(out[i]), want) }
    }
}

