package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCollectIngressIDs_FilePair(t *testing.T) {
    _ = collectIngressIDs(workspace.Workspace{}, ".")
}

