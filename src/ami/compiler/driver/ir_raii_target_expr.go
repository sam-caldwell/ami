package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// raiiTargetFromExpr mirrors sem.releaseTargetFromExpr but is local to avoid exporting internals.
func raiiTargetFromExpr(e ast.Expr) (string, astPos) {
    switch v := e.(type) {
    case *ast.CallExpr:
        return raiiTargetFromCall(v)
    default:
        return "", astPos{}
    }
}

