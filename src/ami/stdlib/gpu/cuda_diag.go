package gpu

import "github.com/sam-caldwell/ami/src/schemas/diag"

// CudaDiag formats a diag for CUDA operations.
func CudaDiag(op string, err error) diag.Record { return BuildDiag("cuda", op, err) }
