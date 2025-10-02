package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func decoArgText(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Name
    case *ast.StringLit:
        return v.Value
    case *ast.NumberLit:
        return v.Text
    default:
        return ""
    }
}

