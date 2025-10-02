package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestFindPackageByRootKey_FilePair(t *testing.T) {
    ws := &workspace.Workspace{}
    _ = findPackageByRootKey(ws, "main")
}

