package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// maybeEmitMethodRecv detects method-form calls and emits a receiver projection prior to the call.
func maybeEmitMethodRecv(st *lowerState, c *ast.CallExpr, out *[]ir.Instruction) {
    if st == nil || c == nil { return }
    if len(c.Args) != 0 { return }
    name := c.Name
    if name == "" { return }
    if !strings.Contains(name, ".") { return }
    last := strings.LastIndex(name, ".")
    recv := name[:last]
    method := name[last+1:]
    if method == "" || recv == "" { return }
    parts := strings.Split(recv, ".")
    var x ast.Expr = &ast.IdentExpr{Name: parts[0]}
    for i := 1; i < len(parts); i++ { x = &ast.SelectorExpr{X: x, Sel: parts[i]} }
    sel, _ := x.(*ast.SelectorExpr)
    if sel == nil { return }
    if ex, ok := lowerSelectorField(st, sel); ok {
        *out = append(*out, ex)
        if ex.Result != nil { st.methodRecv[name] = irValue{id: ex.Result.ID, typ: ex.Result.Type} }
    }
}

