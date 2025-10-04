//go:build darwin

package gpu

import (
    "encoding/binary"
    "math"
    "testing"
)

const mslUseFloatSlice = `#include <metal_stdlib>
using namespace metal;

kernel void mul2_from_const_slice(device float* out [[buffer(0)]],
                                  constant float* arr [[buffer(1)]],
                                  constant uint& n [[buffer(2)]],
                                  uint tid [[thread_position_in_grid]]) {
    if (tid < n) {
        out[tid] = arr[tid] * 2.0;
    }
}`

func TestMetal_Dispatch_WithSlice_Float32(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslUseFloatSlice); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "mul2_from_const_slice"); if err != nil { t.Fatalf("pipe: %v", err) }

    // Prepare input slice and output buffer
    const n = 16
    in := make([]float32, n)
    for i := 0; i < n; i++ { in[i] = float32(i) + 0.5 }
    outBuf, err := MetalAlloc(n * 4) // bytes for n float32
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }

    grid := [3]uint32{n,1,1}
    tpg  := [3]uint32{16,1,1}
    // Bind: out buffer (device), arr as []float32 via setBytes, n as uint32
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, outBuf, in, uint32(n)); err != nil {
        t.Fatalf("dispatch: %v", err)
    }

    raw := make([]byte, n*4)
    if err := MetalCopyFromDevice(raw, outBuf); err != nil { t.Fatalf("copy from device: %v", err) }
    // Decode float32s (LE)
    got := make([]float32, n)
    for i := 0; i < n; i++ {
        u := binary.LittleEndian.Uint32(raw[i*4: i*4+4])
        got[i] = math.Float32frombits(u)
    }
    for i := 0; i < n; i++ {
        want := in[i] * 2
        if diff := math.Abs(float64(got[i]-want)); diff > 1e-5 {
            t.Fatalf("out[%d]=%g want %g (diff %g)", i, got[i], want, diff)
        }
    }
}

