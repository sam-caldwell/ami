package main

import (
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// deriveEdgeAttrs inspects the step corresponding to fromID and returns edge attrs.
// Recognizes:
// - merge.Buffer(size, policy) → bounded (size>0), delivery (policy→bestEffort for dropOldest/dropNewest; atLeastOnce for block)
// - type(TypeName) → type
func deriveEdgeAttrs(pd *ast.PipelineDecl, fromID string, nameToID map[string]string) map[string]any {
    var idx int
    if len(fromID) >= 2 && fromID[2] == ':' {
        tens := int(fromID[0]-'0')
        ones := int(fromID[1]-'0')
        if 0 <= tens && tens <= 9 && 0 <= ones && ones <= 9 { idx = tens*10 + ones }
    }
    si := -1
    for _, s := range pd.Stmts {
        if _, ok := s.(*ast.StepStmt); ok { si++; if si == idx { return attrsFromStep(s.(*ast.StepStmt)) } }
    }
    return nil
}

