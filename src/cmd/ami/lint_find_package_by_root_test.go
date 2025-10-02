package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestFindPackageByRoot_FilePair(t *testing.T) {
    ws := &workspace.Workspace{}
    _ = findPackageByRoot(ws, "./lib")
}

