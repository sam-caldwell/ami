package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestHasUserMain_FilePair(t *testing.T) {
    _ = hasUserMain(workspace.Workspace{}, ".")
}

