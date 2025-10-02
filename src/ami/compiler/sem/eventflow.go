package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeEventTypeFlow performs a conservative type-flow check across pipeline edges.
// See AnalyzeEventTypeFlowInContext for details.
func AnalyzeEventTypeFlow(f *ast.File) []diag.Record { return AnalyzeEventTypeFlowInContext(f, nil) }

