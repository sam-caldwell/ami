package gpu

import "github.com/sam-caldwell/ami/src/schemas/diag"

// OpenCLDiag formats a diag for OpenCL operations.
func OpenCLDiag(op string, err error) diag.Record { return BuildDiag("opencl", op, err) }
