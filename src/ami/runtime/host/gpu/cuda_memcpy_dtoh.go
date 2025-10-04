package gpu

// CudaMemcpyDtoH copies device->host (stub: unavailable).
func CudaMemcpyDtoH(dst []byte, src Buffer) error {
    if !src.valid || src.backend != "cuda" { return ErrInvalidHandle }
    if len(dst) == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

