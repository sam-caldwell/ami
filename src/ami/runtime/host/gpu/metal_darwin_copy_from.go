//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdbool.h>
#include <stdlib.h>

int AmiMetalCopyFromDevice(int bufId, void* dst, unsigned long n, char** err);
*/
import "C"
import (
    "errors"
    "unsafe"
)

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

