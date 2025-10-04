package gpu

// OpenCLLaunchKernel launches an OpenCL kernel (stub: unavailable).
func OpenCLLaunchKernel(ctx Context, k Kernel, global, local [3]uint64, args ...any) error {
    if !ctx.valid || ctx.backend != "opencl" { return ErrInvalidHandle }
    if !k.valid { return ErrInvalidHandle }
    if global[0] == 0 || global[1] == 0 || global[2] == 0 { return ErrInvalidHandle }
    if local[0] == 0 || local[1] == 0 || local[2] == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

