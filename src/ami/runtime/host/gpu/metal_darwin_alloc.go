//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation

int AmiMetalAlloc(unsigned long n, char** err);
*/
import "C"
import "errors"

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

