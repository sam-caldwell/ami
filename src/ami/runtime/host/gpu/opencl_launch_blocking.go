package gpu

// OpenCLLaunchBlocking wraps OpenCLLaunchKernel with panic-safe blocking semantics.
func OpenCLLaunchBlocking(ctx Context, k Kernel, global, local [3]uint64, args ...any) error {
    return Blocking(func() error { return OpenCLLaunchKernel(ctx, k, global, local, args...) })
}

