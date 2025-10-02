package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestDeriveEdgeAttrs_FilePair(t *testing.T) {
    _ = deriveEdgeAttrs(&ast.PipelineDecl{}, "00:x", map[string]string{})
}

