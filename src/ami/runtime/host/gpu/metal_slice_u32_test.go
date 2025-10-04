//go:build darwin

package gpu

import (
    "encoding/binary"
    "testing"
)

const mslAdd5FromU32Slice = `#include <metal_stdlib>
using namespace metal;

kernel void add5_from_u32_slice(device uint* out [[buffer(0)]],
                                constant uint* arr [[buffer(1)]],
                                constant uint& n [[buffer(2)]],
                                uint tid [[thread_position_in_grid]]) {
    if (tid < n) {
        out[tid] = arr[tid] + 5u;
    }
}`

func TestMetal_Dispatch_WithSlice_Uint32(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslAdd5FromU32Slice); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "add5_from_u32_slice"); if err != nil { t.Fatalf("pipe: %v", err) }

    const n = 32
    in := make([]uint32, n)
    for i := 0; i < n; i++ { in[i] = uint32(i) }
    outBuf, err := MetalAlloc(n * 4)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }

    grid := [3]uint32{n,1,1}
    tpg  := [3]uint32{16,1,1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, outBuf, in, uint32(n)); err != nil {
        t.Fatalf("dispatch: %v", err)
    }
    raw := make([]byte, n*4)
    if err := MetalCopyFromDevice(raw, outBuf); err != nil { t.Fatalf("copy from device: %v", err) }
    for i := 0; i < n; i++ {
        got := binary.LittleEndian.Uint32(raw[i*4 : i*4+4])
        want := uint32(i+5)
        if got != want { t.Fatalf("out[%d]=%d want %d", i, got, want) }
    }
}

