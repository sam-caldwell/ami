package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// inferLocalExprTypeWithSigs extends local inference to consult known function
// signatures for call expressions, returning the first result type when known.
func inferLocalExprTypeWithSigs(env map[string]string, sigs map[string][]string, e ast.Expr) string {
    switch v := e.(type) {
    case *ast.CallExpr:
        if rs, ok := sigs[v.Name]; ok && len(rs) > 0 && rs[0] != "" { return rs[0] }
        return "any"
    default:
        return inferLocalExprType(env, e)
    }
}

