package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestCollectMutateWrappedReleases_FilePair(t *testing.T) {
    f := &ast.File{}
    _ = collectMutateWrappedReleases(f)
}

