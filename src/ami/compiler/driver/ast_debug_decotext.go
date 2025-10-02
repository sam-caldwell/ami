package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// decoExprText is a tiny helper to stringify AST expressions for decorator args.
// Keep this aligned with parser's exprText for consistency in debug outputs.
func decoExprText(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Name
    case *ast.StringLit:
        return v.Value
    case *ast.NumberLit:
        return v.Text
    case *ast.SelectorExpr:
        left := decoExprText(v.X)
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

