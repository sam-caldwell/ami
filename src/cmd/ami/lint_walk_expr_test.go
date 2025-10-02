package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestWalkExpr_FilePair(t *testing.T) {
    var e ast.Expr
    walkExpr(e, func(ast.Expr){})
}

