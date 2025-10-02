package gpu

// CudaFree frees device memory (stub validation/unavailable).
func CudaFree(buf Buffer) error {
    if !buf.valid { return ErrInvalidHandle }
    if buf.backend != "cuda" { return ErrInvalidHandle }
    return ErrUnavailable
}

