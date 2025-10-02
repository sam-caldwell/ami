package gpu

import (
    "time"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// BuildDiag constructs a diag.Record for a GPU error path.
func BuildDiag(backend, op string, err error) diag.Record {
    return diag.Record{
        Timestamp: time.Now().UTC(),
        Level:     diag.Error,
        Code:      "E_GPU_" + backend + "_" + op,
        Message:   Explain(backend, op, err),
    }
}
