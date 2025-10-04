package gpu

import (
    "testing"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

func TestGPU_Diag_Builders_MapCodesAndMessages(t *testing.T) {
    r := CudaDiag("alloc", ErrInvalidHandle)
    if r.Code != "E_GPU_INVALID_HANDLE" || r.Level != diag.Error {
        t.Fatalf("cuda diag mismatch: %+v", r)
    }
    if r.Message != "gpu/cuda alloc: invalid handle" { t.Fatalf("msg: %q", r.Message) }

    r = OpenCLDiag("launch", ErrUnavailable)
    if r.Code != "E_GPU_UNAVAILABLE" || r.Level != diag.Error { t.Fatalf("opencl diag mismatch: %+v", r) }
    if r.Message != "gpu/opencl launch: backend unavailable" { t.Fatalf("msg: %q", r.Message) }

    r = MetalDiag("compile", ErrUnimplemented)
    if r.Code != "E_GPU_UNIMPLEMENTED" { t.Fatalf("metal code: %s", r.Code) }
    if r.Message != "gpu/metal compile: unimplemented" { t.Fatalf("msg: %q", r.Message) }
}

