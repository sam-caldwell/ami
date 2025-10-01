//go:build !darwin

package gpu

import "testing"

func TestMetal_Stubs_On_NonDarwin(t *testing.T) {
    if MetalAvailable() { t.Fatalf("MetalAvailable() should be false on non-darwin stub") }
    if d := MetalDevices(); d != nil && len(d) != 0 { t.Fatalf("MetalDevices() expected empty; got %v", d) }
    if _, err := MetalCreateContext(Device{}); err != ErrUnavailable {
        t.Fatalf("MetalCreateContext expected ErrUnavailable; got %v", err)
    }
    if err := MetalDestroyContext(Context{}); err != ErrInvalidHandle {
        t.Fatalf("MetalDestroyContext expected ErrInvalidHandle for zero handle; got %v", err)
    }
    if _, err := MetalCompileLibrary("src"); err != ErrUnavailable { t.Fatalf("MetalCompileLibrary expected ErrUnavailable; got %v", err) }
    if _, err := MetalCreatePipeline(Library{}, "name"); err != ErrUnavailable { t.Fatalf("MetalCreatePipeline expected ErrUnavailable; got %v", err) }
    if err := MetalDispatch(Context{}, Pipeline{}, [3]uint32{1,1,1}, [3]uint32{1,1,1}); err != ErrUnavailable { t.Fatalf("MetalDispatch expected ErrUnavailable; got %v", err) }
    if _, err := MetalAlloc(256); err != ErrUnavailable { t.Fatalf("MetalAlloc expected ErrUnavailable; got %v", err) }
    if err := MetalFree(Buffer{}); err != ErrInvalidHandle { t.Fatalf("MetalFree expected ErrInvalidHandle for zero handle; got %v", err) }
    if err := MetalCopyToDevice(Buffer{}, []byte("hi")); err != ErrUnavailable { t.Fatalf("MetalCopyToDevice expected ErrUnavailable; got %v", err) }
    if err := MetalCopyFromDevice(make([]byte, 2), Buffer{}); err != ErrUnavailable { t.Fatalf("MetalCopyFromDevice expected ErrUnavailable; got %v", err) }
}

