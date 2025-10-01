//go:build darwin

package gpu

import "testing"

const mslIncFromU8Slice = `#include <metal_stdlib>
using namespace metal;

kernel void inc_from_u8_slice(device uchar* out [[buffer(0)]],
                              constant uchar* arr [[buffer(1)]],
                              constant uint& n [[buffer(2)]],
                              uint tid [[thread_position_in_grid]]) {
    if (tid < n) {
        out[tid] = arr[tid] + (uchar)1;
    }
}`

func TestMetal_Dispatch_WithSlice_Uint8(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslIncFromU8Slice); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "inc_from_u8_slice"); if err != nil { t.Fatalf("pipe: %v", err) }

    const n = 32
    in := make([]uint8, n)
    for i := 0; i < n; i++ { in[i] = uint8(i) }
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
        want := uint8(i+1)
        if out[i] != byte(want) { t.Fatalf("out[%d]=%d want %d", i, out[i], want) }
    }
}

