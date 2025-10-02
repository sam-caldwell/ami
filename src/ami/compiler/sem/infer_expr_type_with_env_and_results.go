package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// inferExprTypeWithEnvAndResults attempts to deduce argument types using local env
// and, when the expression is a call, by consulting known function result types
// (only the first result is considered for scalar param positions).
func inferExprTypeWithEnvAndResults(e ast.Expr, vars map[string]string, results map[string][]string) string {
    switch v := e.(type) {
    case *ast.CallExpr:
        if results != nil {
            if rs, ok := results[v.Name]; ok && len(rs) > 0 && rs[0] != "" { return rs[0] }
        }
        return "any"
    default:
        return inferExprTypeWithVars(e, vars)
    }
}

