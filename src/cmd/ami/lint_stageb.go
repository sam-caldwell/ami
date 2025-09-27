package main

import (
    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// lintStageB is a placeholder for parser/semantics-backed rules.
// It currently returns no diagnostics until the frontend is available.
func lintStageB(dir string, ws *workspace.Workspace, t RuleToggles) []diag.Record {
    // TODO: integrate parser/semantics once available and drive rules based on t.
    return nil
}

