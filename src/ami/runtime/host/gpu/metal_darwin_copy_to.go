//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdbool.h>
#include <stdlib.h>

int AmiMetalCopyToDevice(int bufId, const void* src, unsigned long n, char** err);
*/
import "C"
import (
    "errors"
    "unsafe"
)

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
