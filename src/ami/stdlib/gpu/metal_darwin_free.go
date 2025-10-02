//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation

void AmiMetalFreeBuffer(int bufId);
*/
import "C"

func MetalFree(buf Buffer) error {
    if buf.backend != "metal" || !buf.valid || buf.bufId <= 0 { return ErrInvalidHandle }
    C.AmiMetalFreeBuffer(C.int(buf.bufId))
    return nil
}

