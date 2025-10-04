package gpu

// CudaMemcpyHtoD copies host->device (stub: unavailable).
func CudaMemcpyHtoD(dst Buffer, src []byte) error {
    if !dst.valid || dst.backend != "cuda" { return ErrInvalidHandle }
    if len(src) == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

