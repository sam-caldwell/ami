package main

import (
    "strings"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// collectUsedIdents returns a set of identifier names used as top-level prefixes in expressions.
func collectUsedIdents(f *ast.File) map[string]bool {
    used := map[string]bool{}
    walkExprs(f, func(e ast.Expr) {
        switch n := e.(type) {
        case *ast.IdentExpr:
            used[n.Name] = true
        case *ast.CallExpr:
            // Split qualified name on '.' and take first segment
            name := n.Name
            if i := strings.IndexByte(name, '.'); i >= 0 { name = name[:i] }
            used[name] = true
        }
    })
    return used
}

