package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func inferRetTypeWithEnv(e ast.Expr, env map[string]string, results map[string][]string) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if t := env[v.Name]; t != "" { return t }
        return "any"
    case *ast.CallExpr:
        if rs, ok := results[v.Name]; ok && len(rs) > 0 && rs[0] != "" { return rs[0] }
        return "any"
    case *ast.NumberLit, *ast.StringLit, *ast.SliceLit, *ast.SetLit, *ast.MapLit:
        return deduceType(e)
    default:
        return inferExprType(env, e)
    }
}

