package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLintStageB_FilePair(t *testing.T) {
    var ws *workspace.Workspace
    var tgl RuleToggles
    _ = lintStageB(".", ws, tgl)
}

