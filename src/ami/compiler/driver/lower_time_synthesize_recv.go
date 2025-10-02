package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// synthesizeMethodRecvArg extracts receiver value for method-form calls.
func synthesizeMethodRecvArg(st *lowerState, fullName string) (ir.Value, bool) {
    if st == nil || fullName == "" { return ir.Value{}, false }
    if st.methodRecv != nil {
        if v, ok := st.methodRecv[fullName]; ok { return ir.Value{ID: v.id, Type: v.typ}, true }
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
    return ir.Value{ID: ex.Result.ID, Type: "int64"}, true
}

