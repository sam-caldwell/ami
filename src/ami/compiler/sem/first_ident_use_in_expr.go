package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// firstIdentUseInExpr returns the name of the first identifier used in expression `e`
// that appears in the `released` set. Returns empty string when none.
func firstIdentUseInExpr(e ast.Expr, released map[string]bool) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if released[v.Name] { return v.Name }
        return ""
    case *ast.CallExpr:
        for _, a := range v.Args { if n := firstIdentUseInExpr(a, released); n != "" { return n } }
        return ""
    case *ast.UnaryExpr:
        return firstIdentUseInExpr(v.X, released)
    case *ast.BinaryExpr:
        if n := firstIdentUseInExpr(v.X, released); n != "" { return n }
        return firstIdentUseInExpr(v.Y, released)
    default:
        return ""
    }
}

