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
            // If destination variable is Owned and RHS is a literal, wrap via owned_new before assign
            if st != nil && st.varTypes != nil {
                if dtype := st.varTypes[v.Name]; dtype == "Owned" || (len(dtype) >= 6 && dtype[:6] == "Owned<") {
                    // lower RHS expr for data
                    if ex, ok := lowerExpr(st, v.Value); ok {
                        if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                        // determine literal length from AST
                        var length int
                        switch lit := v.Value.(type) {
                        case *ast.StringLit:
                            length = len(lit.Value)
                        case *ast.SliceLit:
                            length = len(lit.Elems)
                        default:
                            length = -1
                        }
                        if length >= 0 {
                            // materialize length
                            lenID := st.newTemp()
                            lres := &ir.Value{ID: lenID, Type: "int64"}
                            out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", length), Result: lres})
                            // call owned_new
                            var data ir.Value
                            if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                            hid := st.newTemp()
                            hres := &ir.Value{ID: hid, Type: "ptr"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, {ID: lenID, Type: "int64"}}, Result: hres})
                            // assign handle
                            out = append(out, ir.Assign{DestID: v.Name, Src: *hres})
                            break
                        }
                    }
                }
            }
            // fallback to simple assign
            out = append(out, lowerStmtAssign(st, v))
        case *ast.ReturnStmt:
            // Materialize return expressions so literals/ops appear as EXPR before RETURN
            var vals []ir.Value
            for i, e := range v.Results {
                if ex, ok := lowerExpr(st, e); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    // If returning Owned and expr is literal, wrap into handle
                    if st != nil && st.currentFn != "" {
                        if rts, ok2 := st.funcResults[st.currentFn]; ok2 && i < len(rts) {
                            rt := rts[i]
                            if rt == "Owned" || (len(rt) >= 6 && rt[:6] == "Owned<") {
                                var length int = -1
                                switch lit := e.(type) {
                                case *ast.StringLit:
                                    length = len(lit.Value)
                                case *ast.SliceLit:
                                    length = len(lit.Elems)
                                }
                                if length >= 0 {
                                    // emit length literal
                                    lenID := st.newTemp()
                                    lres := &ir.Value{ID: lenID, Type: "int64"}
                                    out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", length), Result: lres})
                                    // call owned_new(data,len)
                                    hid := st.newTemp()
                                    hres := &ir.Value{ID: hid, Type: "ptr"}
                                    var data ir.Value
                                    if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, {ID: lenID, Type: "int64"}}, Result: hres})
                                    vals = append(vals, *hres)
                                    continue
                                }
                            }
                        }
                    }
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
                    // Call zeroize+free via runtime helper
                    var argv ir.Value
                    if exArg.Result != nil { argv = *exArg.Result } else { argv = ir.Value{ID: "", Type: "ptr"} }
                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{argv}})
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
