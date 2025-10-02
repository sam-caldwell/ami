package gpu

// CudaCreateContext creates a CUDA context (stub: unavailable).
func CudaCreateContext(dev Device) (Context, error) {
    if dev.Backend != "cuda" || dev.ID < 0 {
        return Context{}, ErrInvalidHandle
    }
    return Context{}, ErrUnavailable
}

