package gpu

// CudaAlloc allocates device memory (stub: unavailable).
func CudaAlloc(n int) (Buffer, error) {
    if n <= 0 { return Buffer{}, ErrInvalidHandle }
    return Buffer{}, ErrUnavailable
}

