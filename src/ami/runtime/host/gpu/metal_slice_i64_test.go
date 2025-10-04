//go:build darwin

package gpu

import (
    "encoding/binary"
    "math"
    "testing"
)

const mslMul3FromI64Slice = `#include <metal_stdlib>
using namespace metal;

kernel void mul3_from_i64_slice(device long* out [[buffer(0)]],
                                constant long* arr [[buffer(1)]],
                                constant uint& n [[buffer(2)]],
                                uint tid [[thread_position_in_grid]]) {
    if (tid < n) {
        out[tid] = arr[tid] * 3;
    }
}`

func TestMetal_Dispatch_WithSlice_Int64(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslMul3FromI64Slice); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "mul3_from_i64_slice"); if err != nil { t.Fatalf("pipe: %v", err) }

    const n = 16
    in := make([]int64, n)
    for i := 0; i < n; i++ { in[i] = int64(i) - 8 }
    outBuf, err := MetalAlloc(n * 8)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }

    grid := [3]uint32{n,1,1}
    tpg  := [3]uint32{16,1,1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, outBuf, in, uint32(n)); err != nil {
        t.Fatalf("dispatch: %v", err)
    }
    raw := make([]byte, n*8)
    if err := MetalCopyFromDevice(raw, outBuf); err != nil { t.Fatalf("copy from device: %v", err) }
    got := make([]int64, n)
    for i := 0; i < n; i++ {
        u := binary.LittleEndian.Uint64(raw[i*8: i*8+8])
        got[i] = int64(u)
    }
    for i := 0; i < n; i++ {
        want := in[i] * 3
        if math.Abs(float64(got[i]-want)) > 0 { t.Fatalf("out[%d]=%d want %d", i, got[i], want) }
    }
}

