package main

import ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func walkExpr(e ast.Expr, fn func(ast.Expr)) {
    if e == nil { return }
    fn(e)
    switch n := e.(type) {
    case *ast.CallExpr:
        for _, a := range n.Args { walkExpr(a, fn) }
    case *ast.BinaryExpr:
        if n.X != nil { walkExpr(n.X, fn) }
        if n.Y != nil { walkExpr(n.Y, fn) }
    case *ast.SelectorExpr:
        if n.X != nil { walkExpr(n.X, fn) }
    }
}

