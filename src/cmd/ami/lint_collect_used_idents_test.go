package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestCollectUsedIdents_FilePair(t *testing.T) {
    f := &ast.File{}
    _ = collectUsedIdents(f)
}

