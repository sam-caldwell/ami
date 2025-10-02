package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestWalkStmts_FilePair(t *testing.T) {
    f := &ast.File{}
    walkStmts(f, func(ast.Stmt){})
}

