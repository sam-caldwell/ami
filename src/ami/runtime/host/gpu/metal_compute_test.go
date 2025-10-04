//go:build darwin

package gpu

import (
    "testing"
)

const mslAddOne = `#include <metal_stdlib>
using namespace metal;

kernel void add_one(device uchar* buf [[buffer(0)]],
                    uint tid [[thread_position_in_grid]]) {
    buf[tid] = buf[tid] + (uchar)1;
}`

func TestMetal_Compile_Pipeline_Dispatch_AddOne(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices()
    if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0])
    if err != nil { t.Fatalf("MetalCreateContext: %v", err) }
    lib, err := MetalCompileLibrary(mslAddOne)
    if err != nil { t.Fatalf("MetalCompileLibrary: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "add_one")
    if err != nil { t.Fatalf("MetalCreatePipeline: %v", err) }
    const n = 64
    buf, err := MetalAlloc(n)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }
    host := make([]byte, n)
    for i := 0; i < n; i++ { host[i] = byte(i) }
    if err := MetalCopyToDevice(buf, host); err != nil { t.Fatalf("MetalCopyToDevice: %v", err) }
    grid := [3]uint32{n, 1, 1}
    tpg := [3]uint32{16, 1, 1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, buf); err != nil {
        t.Fatalf("MetalDispatchBlocking: %v", err)
    }
    out := make([]byte, n)
    if err := MetalCopyFromDevice(out, buf); err != nil { t.Fatalf("MetalCopyFromDevice: %v", err) }
    for i := 0; i < n; i++ {
        want := byte(i + 1)
        if out[i] != want { t.Fatalf("out[%d]=%d want %d", i, out[i], want) }
    }
}

