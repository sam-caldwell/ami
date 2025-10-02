package gpu

import "github.com/sam-caldwell/ami/src/schemas/diag"

// BuildDiag constructs a diag.Error code string for a GPU error path.
func BuildDiag(backend, op string, err error) diag.Error {
    msg := Explain(backend, op, err)
    return diag.Error{Code: "E_GPU_" + backend + "_" + op, Message: msg}
}

