package gpu

import "testing"

const mslNoop = `#include <metal_stdlib>
using namespace metal;
kernel void k0(device uint8_t* out [[ buffer(0) ]], uint32_t i [[ thread_position_in_grid ]]) {
  out[i] = 0;
}`

func TestMetal_Release_LibraryPipelineBuffer(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    devs := MetalDevices(); if len(devs) == 0 { t.Skip("no devices") }

    ctx, err := MetalCreateContext(devs[0])
    if err != nil { t.Fatalf("MetalCreateContext: %v", err) }
    lib, err := MetalCompileLibrary(mslNoop)
    if err != nil { t.Fatalf("MetalCompileLibrary: %v", err) }
    pipe, err := MetalCreatePipeline(lib, "k0")
    if err != nil { t.Fatalf("MetalCreatePipeline: %v", err) }
    buf, err := MetalAlloc(16)
    if err != nil { t.Fatalf("MetalAlloc: %v", err) }
    buf2, err := MetalAlloc(8)
    if err != nil { t.Fatalf("MetalAlloc(2): %v", err) }

    // Exercise free/release to cover cgo helpers; prefer direct MetalFree
    if err := MetalFree(buf); err != nil { t.Fatalf("MetalFree: %v", err) }
    if err := buf2.Release(); err != nil { t.Fatalf("buffer Release: %v", err) }
    if err := pipe.Release(); err != nil { t.Fatalf("pipeline Release: %v", err) }
    if err := lib.Release(); err != nil { t.Fatalf("library Release: %v", err) }
    if err := ctx.Release(); err != nil { t.Fatalf("context Release: %v", err) }
}
