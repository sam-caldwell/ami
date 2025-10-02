package main

import ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// walkExprs invokes fn for every expression node reachable from functions/pipelines.
func walkExprs(f *ast.File, fn func(ast.Expr)) {
    walkStmts(f, func(s ast.Stmt) {
        switch n := s.(type) {
        case *ast.ExprStmt:
            if n.X != nil { walkExpr(n.X, fn) }
        case *ast.AssignStmt:
            if n.Value != nil { walkExpr(n.Value, fn) }
        case *ast.VarDecl:
            if n.Init != nil { walkExpr(n.Init, fn) }
        case *ast.DeferStmt:
            if n.Call != nil { walkExpr(n.Call, fn) }
        case *ast.ReturnStmt:
            for _, e := range n.Results { walkExpr(e, fn) }
        }
    })
}

