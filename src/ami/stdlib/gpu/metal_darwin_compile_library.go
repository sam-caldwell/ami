//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdlib.h>

int AmiMetalCompileLibrary(const char* src, char** err);
*/
import "C"
import (
    "errors"
    "unsafe"
)

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

