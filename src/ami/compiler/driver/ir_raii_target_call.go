package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// raiiTargetFromCall mirrors sem.releaseTargetFromCall.
func raiiTargetFromCall(c *ast.CallExpr) (string, astPos) {
    if c == nil { return "", astPos{} }
    if c.Name == "release" {
        if len(c.Args) >= 1 {
            if id, ok := c.Args[0].(*ast.IdentExpr); ok { return id.Name, astPos{Line: c.NamePos.Line, Column: c.NamePos.Column, Offset: c.NamePos.Offset} }
        }
        return "", astPos{Line: c.NamePos.Line, Column: c.NamePos.Column, Offset: c.NamePos.Offset}
    }
    if c.Name == "mutate" && len(c.Args) == 1 {
        if inner, ok := c.Args[0].(*ast.CallExpr); ok && inner.Name == "release" {
            if len(inner.Args) >= 1 {
                if id, ok := inner.Args[0].(*ast.IdentExpr); ok { return id.Name, astPos{Line: c.NamePos.Line, Column: c.NamePos.Column, Offset: c.NamePos.Offset} }
            }
        }
    }
    return "", astPos{Line: c.NamePos.Line, Column: c.NamePos.Column, Offset: c.NamePos.Offset}
}

