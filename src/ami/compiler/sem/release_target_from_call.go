package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// releaseTargetFromCall inspects a call expression for release(x) or mutate(release(x)).
func releaseTargetFromCall(c *ast.CallExpr) (string, source.Position) {
    if c == nil { return "", source.Position{} }
    if c.Name == "release" {
        if len(c.Args) >= 1 {
            if id, ok := c.Args[0].(*ast.IdentExpr); ok { return id.Name, c.NamePos }
        }
        return "", c.NamePos
    }
    if c.Name == "mutate" && len(c.Args) == 1 {
        if inner, ok := c.Args[0].(*ast.CallExpr); ok && inner.Name == "release" {
            if len(inner.Args) >= 1 {
                if id, ok := inner.Args[0].(*ast.IdentExpr); ok { return id.Name, c.NamePos }
            }
        }
    }
    return "", c.NamePos
}

