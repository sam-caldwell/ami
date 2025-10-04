package gpu

// MetalDispatchBlocking wraps MetalDispatch with panic-safe blocking semantics.
func MetalDispatchBlocking(ctx Context, p Pipeline, grid, threadsPerGroup [3]uint32, args ...any) error {
    return Blocking(func() error { return MetalDispatch(ctx, p, grid, threadsPerGroup, args...) })
}

