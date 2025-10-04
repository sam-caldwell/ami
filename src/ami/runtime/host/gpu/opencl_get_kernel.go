package gpu

import "strings"

// OpenCLGetKernel retrieves a kernel handle (stub: unavailable).
func OpenCLGetKernel(prog Program, name string) (Kernel, error) {
    if !prog.valid { return Kernel{}, ErrInvalidHandle }
    if strings.TrimSpace(name) == "" { return Kernel{}, ErrInvalidHandle }
    return Kernel{}, ErrUnavailable
}

