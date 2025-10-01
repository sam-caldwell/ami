//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdbool.h>
#include <stdlib.h>

// Forward decls implemented in metal_darwin.m
bool AmiMetalAvailable(void);
int AmiMetalDeviceCount(void);
char* AmiMetalDeviceNameAt(int idx);
void AmiMetalFreeCString(char* p);

int AmiMetalContextCreate(int devIndex, char** err);
void AmiMetalContextDestroy(int ctxId);
int AmiMetalCompileLibrary(const char* src, char** err);
int AmiMetalCreatePipeline(int libId, const char* name, char** err);
int AmiMetalAlloc(unsigned long n, char** err);
int AmiMetalCopyToDevice(int bufId, const void* src, unsigned long n, char** err);
int AmiMetalCopyFromDevice(int bufId, void* dst, unsigned long n, char** err);
int AmiMetalDispatch(int ctxId, int pipeId,
                     unsigned int gx, unsigned int gy, unsigned int gz,
                     unsigned int tx, unsigned int ty, unsigned int tz,
                     const int* bufIds, int bufCount, char** err);
void AmiMetalFreeBuffer(int bufId);
*/
import "C"
import (
    "errors"
    "fmt"
    "unsafe"
)

// MetalAvailable reports whether a Metal device is present.
func MetalAvailable() bool { return bool(C.AmiMetalAvailable()) }

// MetalDevices enumerates Metal devices by index with names.
func MetalDevices() []Device {
    n := int(C.AmiMetalDeviceCount())
    if n <= 0 { return nil }
    out := make([]Device, 0, n)
    for i := 0; i < n; i++ {
        nameC := C.AmiMetalDeviceNameAt(C.int(i))
        name := ""
        if nameC != nil {
            name = C.GoString(nameC)
            C.AmiMetalFreeCString(nameC)
        }
        out = append(out, Device{Backend: "metal", ID: i, Name: name})
    }
    return out
}

// MetalCreateContext uses device index to create a context (device+queue).
func MetalCreateContext(dev Device) (Context, error) {
    if !MetalAvailable() { return Context{}, ErrUnavailable }
    var cerr *C.char
    id := int(C.AmiMetalContextCreate(C.int(dev.ID), &cerr))
    if id <= 0 {
        if cerr != nil { return Context{}, errors.New(C.GoString(cerr)) }
        return Context{}, ErrUnavailable
    }
    return Context{backend: "metal", valid: true, ctxId: id}, nil
}

// MetalDestroyContext invalidates the context.
func MetalDestroyContext(ctx Context) error {
    if ctx.backend != "metal" || !ctx.valid || ctx.ctxId <= 0 { return ErrInvalidHandle }
    C.AmiMetalContextDestroy(C.int(ctx.ctxId))
    return nil
}

// MetalCompileLibrary compiles an in-memory MSL source on default device.
func MetalCompileLibrary(src string) (Library, error) {
    csrc := C.CString(src)
    defer C.free(unsafe.Pointer(csrc))
    var cerr *C.char
    id := int(C.AmiMetalCompileLibrary(csrc, &cerr))
    if id <= 0 {
        if cerr != nil { return Library{}, errors.New(C.GoString(cerr)) }
        return Library{}, ErrUnavailable
    }
    return Library{valid: true, libId: id}, nil
}

// MetalCreatePipeline creates a compute pipeline from a library and function name.
func MetalCreatePipeline(lib Library, name string) (Pipeline, error) {
    if !lib.valid || lib.libId <= 0 { return Pipeline{}, ErrInvalidHandle }
    cname := C.CString(name)
    defer C.free(unsafe.Pointer(cname))
    var cerr *C.char
    id := int(C.AmiMetalCreatePipeline(C.int(lib.libId), cname, &cerr))
    if id <= 0 {
        if cerr != nil { return Pipeline{}, errors.New(C.GoString(cerr)) }
        return Pipeline{}, ErrUnavailable
    }
    return Pipeline{valid: true, pipeId: id}, nil
}

// MetalAlloc allocates a shared MTLBuffer on default device.
func MetalAlloc(n int) (Buffer, error) {
    if n <= 0 { return Buffer{}, ErrInvalidHandle }
    var cerr *C.char
    id := int(C.AmiMetalAlloc(C.ulong(n), &cerr))
    if id <= 0 {
        if cerr != nil { return Buffer{}, errors.New(C.GoString(cerr)) }
        return Buffer{}, ErrUnavailable
    }
    return Buffer{backend: "metal", n: n, valid: true, bufId: id}, nil
}

func MetalFree(buf Buffer) error {
    if buf.backend != "metal" || !buf.valid || buf.bufId <= 0 { return ErrInvalidHandle }
    C.AmiMetalFreeBuffer(C.int(buf.bufId))
    return nil
}

func MetalCopyToDevice(dst Buffer, src []byte) error {
    if dst.backend != "metal" || !dst.valid || dst.bufId <= 0 { return ErrInvalidHandle }
    if len(src) > dst.n { return ErrInvalidHandle }
    var cerr *C.char
    var p unsafe.Pointer
    if len(src) > 0 { p = unsafe.Pointer(&src[0]) }
    r := int(C.AmiMetalCopyToDevice(C.int(dst.bufId), p, C.ulong(len(src)), &cerr))
    if r != 0 {
        if cerr != nil { return errors.New(C.GoString(cerr)) }
        return ErrUnavailable
    }
    return nil
}

func MetalCopyFromDevice(dst []byte, src Buffer) error {
    if src.backend != "metal" || !src.valid || src.bufId <= 0 { return ErrInvalidHandle }
    if len(dst) > src.n { return ErrInvalidHandle }
    var cerr *C.char
    var p unsafe.Pointer
    if len(dst) > 0 { p = unsafe.Pointer(&dst[0]) }
    r := int(C.AmiMetalCopyFromDevice(C.int(src.bufId), p, C.ulong(len(dst)), &cerr))
    if r != 0 {
        if cerr != nil { return errors.New(C.GoString(cerr)) }
        return ErrUnavailable
    }
    return nil
}

// MetalDispatch binds buffers in order and dispatches, blocking on completion.
func MetalDispatch(ctx Context, p Pipeline, grid, threadsPerGroup [3]uint32, args ...any) error {
    if ctx.backend != "metal" || !ctx.valid || ctx.ctxId <= 0 { return ErrInvalidHandle }
    if !p.valid || p.pipeId <= 0 { return ErrInvalidHandle }
    // Collect buffer IDs from args; only Buffer is supported for now.
    bufIds := make([]C.int, 0, len(args))
    for i, a := range args {
        b, ok := a.(Buffer)
        if !ok { return fmt.Errorf("gpu: MetalDispatch arg %d not Buffer", i) }
        if b.backend != "metal" || !b.valid || b.bufId <= 0 { return ErrInvalidHandle }
        bufIds = append(bufIds, C.int(b.bufId))
    }
    var cerr *C.char
    var pbuf *C.int
    if len(bufIds) > 0 { pbuf = &bufIds[0] }
    r := int(C.AmiMetalDispatch(C.int(ctx.ctxId), C.int(p.pipeId),
        C.uint(grid[0]), C.uint(grid[1]), C.uint(grid[2]),
        C.uint(threadsPerGroup[0]), C.uint(threadsPerGroup[1]), C.uint(threadsPerGroup[2]),
        pbuf, C.int(len(bufIds)), &cerr))
    if r != 0 {
        if cerr != nil { return errors.New(C.GoString(cerr)) }
        return ErrUnavailable
    }
    return nil
}
