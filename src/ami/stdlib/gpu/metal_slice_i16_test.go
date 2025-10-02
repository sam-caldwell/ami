//go:build darwin

package gpu

import (
    "encoding/binary"
    "testing"
)

const mslNegFromI16Slice = `#include <metal_stdlib>
using namespace metal;

kernel void neg_from_i16_slice(device short* out [[buffer(0)]],
                               constant short* arr [[buffer(1)]],
                               constant uint& n [[buffer(2)]],
                               uint tid [[thread_position_in_grid]]) {
    if (tid < n) { out[tid] = -arr[tid]; }
}`

func TestMetal_Dispatch_WithSlice_Int16(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslNegFromI16Slice); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "neg_from_i16_slice"); if err != nil { t.Fatalf("pipe: %v", err) }
    const n = 8
    in := make([]int16, n)
    for i := 0; i < n; i++ { in[i] = int16(i+1) }
    outBuf, err := MetalAlloc(n * 2)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }
    grid := [3]uint32{n,1,1}
    tpg := [3]uint32{8,1,1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, outBuf, in, uint32(n)); err != nil {
        t.Fatalf("dispatch: %v", err)
    }
    raw := make([]byte, n*2)
    if err := MetalCopyFromDevice(raw, outBuf); err != nil { t.Fatalf("copy from device: %v", err) }
    for i := 0; i < n; i++ {
        got := int16(binary.LittleEndian.Uint16(raw[i*2: i*2+2]))
        want := -in[i]
        if got != want { t.Fatalf("out[%d]=%d want %d", i, got, want) }
    }
}

