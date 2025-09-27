package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeAmbiguity reports E_TYPE_AMBIGUOUS for container literals with no
// type arguments and no elements when such syntax exists in this phase.
// Current parser requires explicit type names in slice/set/map literals, so
// this function is a scaffold that returns empty results.
func AnalyzeAmbiguity(f *ast.File) []diag.Record {
    _ = f
    _ = time.Unix(0, 0).UTC()
    return nil
}

