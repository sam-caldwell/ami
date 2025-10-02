package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// releaseTargetFromExpr inspects an expression and returns the released variable name if the
// expression matches release(x) or mutate(release(x)). It also returns the position of the call.
func releaseTargetFromExpr(e ast.Expr) (string, source.Position) {
    switch v := e.(type) {
    case *ast.CallExpr:
        return releaseTargetFromCall(v)
    default:
        return "", source.Position{}
    }
}

