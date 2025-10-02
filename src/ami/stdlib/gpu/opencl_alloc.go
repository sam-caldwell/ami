package gpu

// OpenCLAlloc allocates device memory (stub: unavailable).
func OpenCLAlloc(n int) (Buffer, error) {
    if n <= 0 { return Buffer{}, ErrInvalidHandle }
    return Buffer{}, ErrUnavailable
}

