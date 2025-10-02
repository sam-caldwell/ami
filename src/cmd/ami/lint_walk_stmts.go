package main

import ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// walkStmts invokes fn for every statement in function and pipeline bodies.
func walkStmts(f *ast.File, fn func(ast.Stmt)) {
    for _, d := range f.Decls {
        switch n := d.(type) {
        case *ast.FuncDecl:
            if n != nil && n.Body != nil {
                for _, s := range n.Body.Stmts { fn(s) }
            }
        case *ast.PipelineDecl:
            if n != nil && n.Body != nil {
                for _, s := range n.Body.Stmts { fn(s) }
            }
        }
    }
}

