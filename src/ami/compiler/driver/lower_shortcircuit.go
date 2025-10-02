package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// needsShortCircuit reports whether e contains a conditional or boolean &&/|| that
// should be lowered with control-flow rather than eager evaluation.
func needsShortCircuit(e ast.Expr) bool {
    switch v := e.(type) {
    case *ast.ConditionalExpr:
        return true
    case *ast.BinaryExpr:
        if v.Op == token.And || v.Op == token.Or { return true }
        return needsShortCircuit(v.X) || needsShortCircuit(v.Y)
    case *ast.UnaryExpr:
        return needsShortCircuit(v.X)
    case *ast.CallExpr:
        for _, a := range v.Args { if needsShortCircuit(a) { return true } }
        return false
    case *ast.SelectorExpr:
        return needsShortCircuit(v.X)
    case *ast.SliceLit:
        for _, e2 := range v.Elems { if needsShortCircuit(e2) { return true } }
        return false
    case *ast.SetLit:
        for _, e2 := range v.Elems { if needsShortCircuit(e2) { return true } }
        return false
    case *ast.MapLit:
        for _, kv := range v.Elems { if needsShortCircuit(kv.Key) || needsShortCircuit(kv.Val) { return true } }
        return false
    default:
        return false
    }
}
// lowerValueSC moved to lower_value_sc.go
