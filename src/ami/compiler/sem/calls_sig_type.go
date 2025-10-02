package sem

import (
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

type sig struct{ params, results []string; paramNames []string; paramTypePos []diag.Position }

