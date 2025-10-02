package gpu

import "strings"

// OpenCLBuildProgram builds an OpenCL program (stub: unavailable).
func OpenCLBuildProgram(src string) (Program, error) {
    if strings.TrimSpace(src) == "" { return Program{}, ErrInvalidHandle }
    return Program{}, ErrUnavailable
}

