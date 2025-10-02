package gpu

import "testing"

func TestOpenCL_ProgramKernel_ArgValidation(t *testing.T) {
    if _, err := OpenCLBuildProgram(""); err != ErrInvalidHandle {
        t.Fatalf("OpenCLBuildProgram empty: %v", err)
    }
    if _, err := OpenCLGetKernel(Program{}, "k"); err != ErrInvalidHandle {
        t.Fatalf("OpenCLGetKernel invalid program: %v", err)
    }
    if _, err := OpenCLGetKernel(Program{valid: true}, ""); err != ErrInvalidHandle {
        t.Fatalf("OpenCLGetKernel empty name: %v", err)
    }
}

