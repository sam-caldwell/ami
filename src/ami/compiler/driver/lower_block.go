package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "fmt"
)

// lowerBlock lowers a function body block into a sequence of IR instructions.
func lowerBlock(st *lowerState, b *ast.BlockStmt) []ir.Instruction {
    var out []ir.Instruction
    if b == nil { return out }
    for _, s := range b.Stmts {
        switch v := s.(type) {
        case *ast.VarDecl:
            // Owned ABI: wrap literal initializers into a runtime Owned handle with known length
            if v.Type != "" && (v.Type == "Owned" || (len(v.Type) >= 6 && v.Type[:6] == "Owned<")) && v.Init != nil {
                // lower initializer first
                if ex, ok := lowerExpr(st, v.Init); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    // determine literal length when available
                    var length int
                    switch lit := v.Init.(type) {
                    case *ast.StringLit:
                        length = len(lit.Value)
                    case *ast.SliceLit:
                        length = len(lit.Elems)
                    default:
                        length = 0
                    }
                    // materialize length constant
                    lenTmp := st.newTemp()
                    lres := &ir.Value{ID: lenTmp, Type: "int64"}
                    lz := ir.Expr{Op: fmt.Sprintf("lit:%d", length), Result: lres}
                    out = append(out, lz)
                    // call owned_new(ptr,len)
                    hid := st.newTemp()
                    hres := &ir.Value{ID: hid, Type: "ptr"}
                    var data ir.Value
                    if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, {ID: lenTmp, Type: "int64"}}, Result: hres})
                    // emit var using the owned handle
                    res := ir.Value{ID: v.Name, Type: v.Type}
                    out = append(out, ir.Var{Name: v.Name, Type: v.Type, Init: hres, Result: res})
                    if st != nil && st.varTypes != nil && v.Name != "" { st.varTypes[v.Name] = v.Type }
                    break
                }
            } else {
                out = append(out, lowerStmtVar(st, v))
            }
        case *ast.AssignStmt:
            out = append(out, lowerStmtAssign(st, v))
        case *ast.ReturnStmt:
            // Materialize return expressions so literals/ops appear as EXPR before RETURN
            var vals []ir.Value
            for _, e := range v.Results {
                if ex, ok := lowerExpr(st, e); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    if ex.Result != nil { vals = append(vals, *ex.Result) }
                }
            }
            out = append(out, ir.Return{Values: vals})
        case *ast.DeferStmt:
            out = append(out, lowerStmtDefer(st, v))
        case *ast.ExprStmt:
            // Special-case release(x): emit zeroization call
            if ce, isCall := v.X.(*ast.CallExpr); isCall && ce.Name == "release" && len(ce.Args) >= 1 {
                if exArg, ok := lowerExpr(st, ce.Args[0]); ok {
                    if exArg.Op != "" || exArg.Callee != "" || len(exArg.Args) > 0 { out = append(out, exArg) }
                    // Obtain Owned length via runtime ABI: len = ami_rt_owned_len(x)
                    lenRes := &ir.Value{ID: st.newTemp(), Type: "int64"}
                    var argv ir.Value
                    if exArg.Result != nil { argv = *exArg.Result } else { argv = ir.Value{ID: "", Type: "ptr"} }
                    zlen := ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{argv}, Result: lenRes}
                    out = append(out, zlen)
                    // call zeroize(ptr, len)
                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_zeroize", Args: []ir.Value{argv, {ID: lenRes.ID, Type: "int64"}}})
                }
            } else {
                if e, ok := lowerExpr(st, v.X); ok {
                    out = append(out, e)
                }
            }
        }
    }
    return out
}
