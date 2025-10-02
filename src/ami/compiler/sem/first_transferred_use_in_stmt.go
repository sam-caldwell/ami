package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// firstTransferredUseInStmt scans a statement and returns the first identifier and position
// whose name appears in the transferred set.
func firstTransferredUseInStmt(s ast.Stmt, transferred map[string]bool) (string, source.Position) {
    if s == nil { return "", source.Position{} }
    var check func(e ast.Expr) (string, source.Position)
    check = func(e ast.Expr) (string, source.Position) {
        switch v := e.(type) {
        case *ast.IdentExpr:
            if transferred[v.Name] { return v.Name, v.Pos }
            return "", source.Position{}
        case *ast.CallExpr:
            for _, a := range v.Args { if n, p := check(a); n != "" { return n, p } }
            return "", source.Position{}
        case *ast.UnaryExpr:
            return check(v.X)
        case *ast.BinaryExpr:
            if n, p := check(v.X); n != "" { return n, p }
            return check(v.Y)
        case *ast.SliceLit:
            for _, el := range v.Elems { if n, p := check(el); n != "" { return n, p } }
            return "", source.Position{}
        case *ast.SetLit:
            for _, el := range v.Elems { if n, p := check(el); n != "" { return n, p } }
            return "", source.Position{}
        case *ast.MapLit:
            for _, kv := range v.Elems { if n, p := check(kv.Key); n != "" { return n, p }; if n, p := check(kv.Val); n != "" { return n, p } }
            return "", source.Position{}
        default:
            return "", source.Position{}
        }
    }
    switch v := s.(type) {
    case *ast.ExprStmt:
        if v.X != nil { return check(v.X) }
    case *ast.AssignStmt:
        if v.Value != nil { return check(v.Value) }
    case *ast.VarDecl:
        if v.Init != nil { return check(v.Init) }
    case *ast.DeferStmt:
        if v.Call != nil { return check(v.Call) }
    case *ast.ReturnStmt:
        for _, e := range v.Results { if n, p := check(e); n != "" { return n, p } }
    }
    return "", source.Position{}
}

