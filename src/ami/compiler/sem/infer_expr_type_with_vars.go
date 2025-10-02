package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func inferExprTypeWithVars(e ast.Expr, vars map[string]string) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if t, ok := vars[v.Name]; ok && t != "" { return t }
        return "any"
    case *ast.StringLit:
        return "string"
    case *ast.NumberLit:
        return "int"
    case *ast.SliceLit, *ast.SetLit, *ast.MapLit:
        return deduceType(e)
    case *ast.CallExpr:
        // unknown without inter-procedural lookup here
        return "any"
    default:
        return "any"
    }
}

