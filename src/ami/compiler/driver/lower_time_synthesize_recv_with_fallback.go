package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// synthesizeMethodRecvArgWithFallback extracts the receiver value for a method-form call
// encoded in fullName (e.g., "r.Read", "x.a.b.Scan"). It returns the lowered IR value
// of the receiver. If the receiver's static type cannot be determined, it uses
// fallbackType for the returned IR value type.
func synthesizeMethodRecvArgWithFallback(st *lowerState, fullName, fallbackType string) (ir.Value, bool) {
    if st == nil || fullName == "" { return ir.Value{}, false }
    if st.methodRecv != nil {
        if v, ok := st.methodRecv[fullName]; ok {
            typ := v.typ
            if typ == "" || typ == "any" { typ = fallbackType }
            return ir.Value{ID: v.id, Type: typ}, true
        }
    }
    i := strings.LastIndex(fullName, ".")
    if i <= 0 { return ir.Value{}, false }
    recv := fullName[:i]
    parts := strings.Split(recv, ".")
    if len(parts) == 0 { return ir.Value{}, false }
    var x ast.Expr = &ast.IdentExpr{Name: parts[0]}
    for j := 1; j < len(parts); j++ { x = &ast.SelectorExpr{X: x, Sel: parts[j]} }
    ex, ok := lowerExpr(st, x)
    if !ok || ex.Result == nil { return ir.Value{}, false }
    typ := ex.Result.Type
    if typ == "" || typ == "any" { typ = fallbackType }
    return ir.Value{ID: ex.Result.ID, Type: typ}, true
}

