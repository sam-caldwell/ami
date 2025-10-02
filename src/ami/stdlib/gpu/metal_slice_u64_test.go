//go:build darwin

package gpu

import (
    "encoding/binary"
    "testing"
)

const mslAdd1FromU64Slice = `#include <metal_stdlib>
using namespace metal;

kernel void add1_from_u64_slice(device ulong* out [[buffer(0)]],
                                constant ulong* arr [[buffer(1)]],
                                constant uint& n [[buffer(2)]],
                                uint tid [[thread_position_in_grid]]) {
    if (tid < n) { out[tid] = arr[tid] + 1; }
}`

func TestMetal_Dispatch_WithSlice_Uint64(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslAdd1FromU64Slice); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "add1_from_u64_slice"); if err != nil { t.Fatalf("pipe: %v", err) }
    const n = 8
    in := make([]uint64, n)
    for i := 0; i < n; i++ { in[i] = uint64(i) }
    outBuf, err := MetalAlloc(n * 8)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }
    grid := [3]uint32{n,1,1}
    tpg := [3]uint32{8,1,1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, outBuf, in, uint32(n)); err != nil {
        t.Fatalf("dispatch: %v", err)
    }
    raw := make([]byte, n*8)
    if err := MetalCopyFromDevice(raw, outBuf); err != nil { t.Fatalf("copy from device: %v", err) }
    for i := 0; i < n; i++ {
        got := binary.LittleEndian.Uint64(raw[i*8: i*8+8])
        want := in[i] + 1
        if got != want { t.Fatalf("out[%d]=%d want %d", i, got, want) }
    }
}

