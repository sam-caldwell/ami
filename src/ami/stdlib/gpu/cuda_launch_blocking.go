package gpu

// CudaLaunchBlocking wraps CudaLaunchKernel with panic-safe blocking semantics.
func CudaLaunchBlocking(ctx Context, k Kernel, grid, block [3]uint32, sharedMem uint32, args ...any) error {
    return Blocking(func() error { return CudaLaunchKernel(ctx, k, grid, block, sharedMem, args...) })
}

