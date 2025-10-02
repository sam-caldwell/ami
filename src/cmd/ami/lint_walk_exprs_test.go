package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestWalkExprs_FilePair(t *testing.T) {
    f := &ast.File{}
    walkExprs(f, func(ast.Expr){})
}

