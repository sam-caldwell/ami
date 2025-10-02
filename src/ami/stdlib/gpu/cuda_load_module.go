package gpu

import "strings"

// CudaLoadModule loads a PTX module (stub: unavailable).
func CudaLoadModule(ptx string) (Module, error) {
    if strings.TrimSpace(ptx) == "" { return Module{}, ErrInvalidHandle }
    return Module{}, ErrUnavailable
}

