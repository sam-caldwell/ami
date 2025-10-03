package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestWalkExpr_Recurses_Call_Binary_Selector(t *testing.T) {
    // f(g(x + y).z)
    bx := &ast.BinaryExpr{X: &ast.IdentExpr{Name: "x"}, Y: &ast.IdentExpr{Name: "y"}}
    sel := &ast.SelectorExpr{X: &ast.CallExpr{Name: "g", Args: []ast.Expr{bx}}, Sel: "z"}
    root := &ast.CallExpr{Name: "f", Args: []ast.Expr{sel}}
    count := 0
    walkExpr(root, func(ast.Expr){ count++ })
    if count < 5 { t.Fatalf("expected traversal to visit >=5 nodes, got %d", count) }
}
