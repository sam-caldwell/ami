//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdlib.h>

int AmiMetalContextCreate(int devIndex, char** err);
*/
import "C"
import "errors"

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

