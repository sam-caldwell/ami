package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCollectLocalImportRoots_FilePair(t *testing.T) {
    var ws workspace.Workspace
    var pkg workspace.Package
    _ = collectLocalImportRoots(&ws, &pkg)
}

