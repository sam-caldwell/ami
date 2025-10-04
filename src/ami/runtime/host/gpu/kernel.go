package gpu

// Kernel represents a CUDA kernel or OpenCL kernel.
type Kernel struct{
    valid bool
}

func (k *Kernel) Release() error {
    if k == nil || !k.valid { return ErrInvalidHandle }
    k.valid = false
    return nil
}

