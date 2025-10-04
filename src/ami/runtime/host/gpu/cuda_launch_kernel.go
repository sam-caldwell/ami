package gpu

// CudaLaunchKernel launches a CUDA kernel (stub: unavailable).
func CudaLaunchKernel(ctx Context, k Kernel, grid, block [3]uint32, sharedMem uint32, args ...any) error {
    if !ctx.valid || ctx.backend != "cuda" { return ErrInvalidHandle }
    if !k.valid { return ErrInvalidHandle }
    if grid[0] == 0 || grid[1] == 0 || grid[2] == 0 { return ErrInvalidHandle }
    if block[0] == 0 || block[1] == 0 || block[2] == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

