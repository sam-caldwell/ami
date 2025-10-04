//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation

void AmiMetalContextDestroy(int ctxId);
*/
import "C"

// MetalDestroyContext invalidates the context.
func MetalDestroyContext(ctx Context) error {
    if ctx.backend != "metal" || !ctx.valid || ctx.ctxId <= 0 { return ErrInvalidHandle }
    C.AmiMetalContextDestroy(C.int(ctx.ctxId))
    return nil
}

