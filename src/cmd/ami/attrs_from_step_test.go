package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestAttrsFromStep_FilePair(t *testing.T) {
    _ = attrsFromStep(&ast.StepStmt{})
}

