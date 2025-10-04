//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

// Extended dispatch supporting scalar arguments via setBytes. For each arg index:
// kinds[i] == 0 -> use bufIds[i] with setBuffer; kinds[i] == 1 -> use bytesPtrs[i]/bytesLens[i] with setBytes.
int AmiMetalDispatchEx(int ctxId, int pipeId,
                       unsigned int gx, unsigned int gy, unsigned int gz,
                       unsigned int tx, unsigned int ty, unsigned int tz,
                       const int* kinds, int argCount,
                       const int* bufIds,
                       char** bytesPtrs, unsigned long* bytesLens,
                       char** err);
*/
import "C"
import (
    "errors"
    "fmt"
    "unsafe"
)

// MetalDispatch binds buffers or scalar arguments in order and dispatches, blocking on completion.
// Supported scalars: int32/int64/uint32/uint64/float32/float64. Scalars are passed via setBytes.
func MetalDispatch(ctx Context, p Pipeline, grid, threadsPerGroup [3]uint32, args ...any) error {
    if ctx.backend != "metal" || !ctx.valid || ctx.ctxId <= 0 { return ErrInvalidHandle }
    if !p.valid || p.pipeId <= 0 { return ErrInvalidHandle }
    kinds := make([]C.int, len(args))
    bufIds := make([]C.int, len(args))
    cBytePtrs := make([]*C.char, len(args))
    byteLens := make([]C.ulong, len(args))
    // Track C allocations to free after dispatch
    var cAllocs []unsafe.Pointer
    toBytes := func(v any) (unsafe.Pointer, C.ulong, bool) {
        switch x := v.(type) {
        case []byte:
            if len(x) == 0 { return nil, 0, false }
            p := C.CBytes(x)
            return p, C.ulong(len(x)), true
        case []int8:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x))
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []float32:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 4)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []float64:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 8)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []uint16:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 2)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []int16:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 2)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []uint32:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 4)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []int32:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 4)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []uint64:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 8)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case []int64:
            if len(x) == 0 { return nil, 0, false }
            nbytes := C.ulong(len(x) * 8)
            p := C.malloc(nbytes)
            if p == nil { return nil, 0, false }
            C.memcpy(p, unsafe.Pointer(&x[0]), nbytes)
            return p, nbytes, true
        case int8:
            p := C.CBytes((*[1]byte)(unsafe.Pointer(&x))[:])
            return p, 1, true
        case uint8:
            p := C.CBytes((*[1]byte)(unsafe.Pointer(&x))[:])
            return p, 1, true
        case int16:
            p := C.CBytes((*[2]byte)(unsafe.Pointer(&x))[:])
            return p, 2, true
        case uint16:
            p := C.CBytes((*[2]byte)(unsafe.Pointer(&x))[:])
            return p, 2, true
        case int32:
            p := C.CBytes((*[4]byte)(unsafe.Pointer(&x))[:])
            return p, 4, true
        case uint32:
            p := C.CBytes((*[4]byte)(unsafe.Pointer(&x))[:])
            return p, 4, true
        case int64:
            p := C.CBytes((*[8]byte)(unsafe.Pointer(&x))[:])
            return p, 8, true
        case uint64:
            p := C.CBytes((*[8]byte)(unsafe.Pointer(&x))[:])
            return p, 8, true
        case float32:
            p := C.CBytes((*[4]byte)(unsafe.Pointer(&x))[:])
            return p, 4, true
        case float64:
            p := C.CBytes((*[8]byte)(unsafe.Pointer(&x))[:])
            return p, 8, true
        }
        return nil, 0, false
    }
    for i, a := range args {
        if b, ok := a.(Buffer); ok {
            if b.backend != "metal" || !b.valid || b.bufId <= 0 { return ErrInvalidHandle }
            kinds[i] = 0
            bufIds[i] = C.int(b.bufId)
            continue
        }
        if p, n, ok := toBytes(a); ok {
            kinds[i] = 1
            cBytePtrs[i] = (*C.char)(p)
            byteLens[i] = n
            cAllocs = append(cAllocs, p)
            continue
        }
        return fmt.Errorf("gpu: MetalDispatch arg %d unsupported type", i)
    }
    var cerr *C.char
    var pkinds *C.int
    var pbuf *C.int
    var pbytes **C.char
    var plens *C.ulong
    if len(kinds) > 0 { pkinds = &kinds[0] }
    if len(bufIds) > 0 { pbuf = &bufIds[0] }
    if len(cBytePtrs) > 0 { pbytes = (**C.char)(unsafe.Pointer(&cBytePtrs[0])) }
    if len(byteLens) > 0 { plens = &byteLens[0] }
    r := int(C.AmiMetalDispatchEx(C.int(ctx.ctxId), C.int(p.pipeId),
        C.uint(grid[0]), C.uint(grid[1]), C.uint(grid[2]),
        C.uint(threadsPerGroup[0]), C.uint(threadsPerGroup[1]), C.uint(threadsPerGroup[2]),
        pkinds, C.int(len(kinds)), pbuf, pbytes, plens, &cerr))
    // Free any C allocations for setBytes
    for _, p := range cAllocs { C.free(p) }
    if r != 0 {
        if cerr != nil { return errors.New(C.GoString(cerr)) }
        return ErrUnavailable
    }
    return nil
}

