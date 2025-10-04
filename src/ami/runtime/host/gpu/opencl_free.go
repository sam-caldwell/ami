package gpu

// OpenCLFree frees device memory (stub validation/unavailable).
func OpenCLFree(buf Buffer) error {
    if !buf.valid { return ErrInvalidHandle }
    if buf.backend != "opencl" { return ErrInvalidHandle }
    return ErrUnavailable
}

