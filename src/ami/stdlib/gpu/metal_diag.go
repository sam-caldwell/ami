package gpu

import "github.com/sam-caldwell/ami/src/schemas/diag"

// MetalDiag formats a diag for Metal operations.
func MetalDiag(op string, err error) diag.Record { return BuildDiag("metal", op, err) }
