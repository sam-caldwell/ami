package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLintWorkspace_FilePair(t *testing.T) {
    ws := &workspace.Workspace{}
    _ = lintWorkspace(".", ws)
}

