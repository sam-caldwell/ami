package gpu

import (
    "time"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// BuildDiag constructs a diag.Record for a GPU error path.
func BuildDiag(backend, op string, err error) diag.Record {
    code := "E_GPU_UNKNOWN"
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
    }
}
