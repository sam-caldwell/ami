package gpu

import "strings"

// CudaGetKernel retrieves a kernel handle (stub: unavailable).
func CudaGetKernel(mod Module, name string) (Kernel, error) {
    if !mod.valid { return Kernel{}, ErrInvalidHandle }
    if strings.TrimSpace(name) == "" { return Kernel{}, ErrInvalidHandle }
    return Kernel{}, ErrUnavailable
}

