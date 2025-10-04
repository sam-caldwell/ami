//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdlib.h>

int AmiMetalCreatePipeline(int libId, const char* name, char** err);
*/
import "C"
import (
    "errors"
    "unsafe"
)

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

