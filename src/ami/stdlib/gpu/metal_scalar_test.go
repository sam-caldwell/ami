//go:build darwin

package gpu

import "testing"

const mslAddConst = `#include <metal_stdlib>
using namespace metal;

kernel void add_const(device uchar* buf [[buffer(0)]],
                      constant uchar& c [[buffer(1)]],
                      uint tid [[thread_position_in_grid]]) {
    buf[tid] = buf[tid] + c;
}`

func TestMetal_Dispatch_WithScalar_ConstantArg(t *testing.T) {
    if !MetalAvailable() { t.Skip("Metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }
    ctx, err := MetalCreateContext(devs[0]); if err != nil { t.Fatalf("ctx: %v", err) }
    lib, err := MetalCompileLibrary(mslAddConst); if err != nil { t.Fatalf("lib: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "add_const"); if err != nil { t.Fatalf("pipe: %v", err) }
    const n = 32
    buf, err := MetalAlloc(n); if err != nil { t.Fatalf("alloc: %v", err) }
    host := make([]byte, n)
    for i := 0; i < n; i++ { host[i] = byte(i) }
    if err := MetalCopyToDevice(buf, host); err != nil { t.Fatalf("copy to device: %v", err) }
    add := uint32(5)
    grid := [3]uint32{n,1,1}
    tpg := [3]uint32{16,1,1}
    if err := MetalDispatchBlocking(ctx, pipe, grid, tpg, buf, add); err != nil {
        t.Fatalf("dispatch: %v", err)
    }
    out := make([]byte, n)
    if err := MetalCopyFromDevice(out, buf); err != nil { t.Fatalf("copy from device: %v", err) }
    for i := 0; i < n; i++ {
        want := byte(int(host[i]) + 5)
        if out[i] != want { t.Fatalf("out[%d]=%d want %d", i, out[i], want) }
    }
}

