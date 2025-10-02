package gpu

import (
    "time"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// BuildDiag constructs a diag.Record for a GPU error with deterministic fields.
func BuildDiag(backend, op string, err error) diag.Record {
    code := "E_GPU_ERROR"
    switch err {
    case ErrInvalidHandle:
        code = "E_GPU_INVALID_HANDLE"
    case ErrUnavailable:
        code = "E_GPU_UNAVAILABLE"
    case ErrUnimplemented:
        code = "E_GPU_UNIMPLEMENTED"
    }
    return diag.Record{
        Timestamp: time.Now().UTC(),
        Level:     diag.Error,
        Code:      code,
        Message:   Explain(backend, op, err),
        Package:   "stdlib.gpu",
    }
}

func CudaDiag(op string, err error) diag.Record   { return BuildDiag("cuda", op, err) }
func OpenCLDiag(op string, err error) diag.Record { return BuildDiag("opencl", op, err) }
func MetalDiag(op string, err error) diag.Record  { return BuildDiag("metal", op, err) }

