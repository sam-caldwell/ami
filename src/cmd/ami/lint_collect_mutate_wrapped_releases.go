package main

import (
    "strings"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// collectMutateWrappedReleases returns keys of release calls that are wrapped in mutate().
func collectMutateWrappedReleases(f *ast.File) map[string]bool {
    wrapped := map[string]bool{}
    walkExprs(f, func(e ast.Expr) {
        ce, ok := e.(*ast.CallExpr)
        if !ok { return }
        if strings.EqualFold(ce.Name, "mutate") && len(ce.Args) > 0 {
            if inner, ok := ce.Args[0].(*ast.CallExpr); ok {
                lname := strings.ToLower(inner.Name)
                if lname == "release" || strings.HasSuffix(lname, ".release") {
                    wrapped[callKey("", inner)] = true
                }
            }
        }
    })
    return wrapped
}

