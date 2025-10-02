package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// debugExprText mirrors the simple printer used in debug JSON paths.
func debugExprText(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Name
    case *ast.StringLit:
        return v.Value
    case *ast.NumberLit:
        return v.Text
    case *ast.SelectorExpr:
        left := debugExprText(v.X)
        if left == "" { left = "?" }
        return left + "." + v.Sel
    case *ast.CallExpr:
        if len(v.Args) > 0 { return v.Name + "(…)" }
        return v.Name + "()"
    case *ast.SliceLit:
        return "slice"
    case *ast.SetLit:
        return "set"
    case *ast.MapLit:
        return "map"
    default:
        return ""
    }
}

