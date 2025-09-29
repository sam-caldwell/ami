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
            // Owned ABI: unconditional copy-on-new for Owned variables with initializers.
            if v.Type != "" && (v.Type == "Owned" || (len(v.Type) >= 6 && v.Type[:6] == "Owned<")) && v.Init != nil {
                // lower initializer first
                if ex, ok := lowerExpr(st, v.Init); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    var data ir.Value
                    var lenVal ir.Value
                    // If initializer is a literal we can compute length; otherwise, if it's an Owned handle, query ptr/len from runtime.
                    switch lit := v.Init.(type) {
                    case *ast.StringLit:
                        // literal string: use known length and data pointer from lowered expr
                        length := len(lit.Value)
                        lenTmp := st.newTemp()
                        lres := &ir.Value{ID: lenTmp, Type: "int64"}
                        out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", length), Result: lres})
                        lenVal = *lres
                        if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                    case *ast.SliceLit:
                        length := len(lit.Elems)
                        lenTmp := st.newTemp()
                        lres := &ir.Value{ID: lenTmp, Type: "int64"}
                        out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", length), Result: lres})
                        lenVal = *lres
                        if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                    default:
                        // Non-literal: if lowered expr is an Owned handle, derive data/len via runtime helpers
                        if ex.Result != nil && (ex.Result.Type == "Owned" || (len(ex.Result.Type) >= 6 && ex.Result.Type[:6] == "Owned<")) {
                            src := *ex.Result
                            // query pointer
                            ptmp := st.newTemp()
                            pres := &ir.Value{ID: ptmp, Type: "ptr"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{src}, Result: pres})
                            data = *pres
                            // query length
                            ltmp := st.newTemp()
                            lres := &ir.Value{ID: ltmp, Type: "int64"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{src}, Result: lres})
                            lenVal = *lres
                        } else {
                            // Fallback: no copy wrapper; use generic var lowering
                            out = append(out, lowerStmtVar(st, v))
                            break
                        }
                    }
                    // call owned_new(ptr,len)
                    hid := st.newTemp()
                    hres := &ir.Value{ID: hid, Type: "Owned"}
                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, lenVal}, Result: hres})
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
            // If destination variable is Owned, wrap RHS via owned_new before assign (copy-on-new)
            if st != nil && st.varTypes != nil {
                if dtype := st.varTypes[v.Name]; dtype == "Owned" || (len(dtype) >= 6 && dtype[:6] == "Owned<") {
                    // lower RHS expr
                    if ex, ok := lowerExpr(st, v.Value); ok {
                        if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                        var data ir.Value
                        var lenVal ir.Value
                        // literal lengths when available
                        switch lit := v.Value.(type) {
                        case *ast.StringLit:
                            l := len(lit.Value)
                            lenID := st.newTemp()
                            lres := &ir.Value{ID: lenID, Type: "int64"}
                            out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", l), Result: lres})
                            lenVal = *lres
                            if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                        case *ast.SliceLit:
                            l := len(lit.Elems)
                            lenID := st.newTemp()
                            lres := &ir.Value{ID: lenID, Type: "int64"}
                            out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", l), Result: lres})
                            lenVal = *lres
                            if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                        default:
                            // Non-literal: if RHS is Owned, derive ptr/len via runtime
                            if ex.Result != nil && (ex.Result.Type == "Owned" || (len(ex.Result.Type) >= 6 && ex.Result.Type[:6] == "Owned<")) {
                                src := *ex.Result
                                // ptr
                                ptmp := st.newTemp(); pres := &ir.Value{ID: ptmp, Type: "ptr"}
                                out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{src}, Result: pres})
                                data = *pres
                                // len
                                ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                                out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{src}, Result: lres})
                                lenVal = *lres
                            } else {
                                // Fallback
                                out = append(out, lowerStmtAssign(st, v))
                                break
                            }
                        }
                        // call owned_new and assign
                        hid := st.newTemp(); hres := &ir.Value{ID: hid, Type: "Owned"}
                        out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, lenVal}, Result: hres})
                        out = append(out, ir.Assign{DestID: v.Name, Src: *hres})
                        break
                    }
                }
            }
            // fallback to simple assign if not handled
            out = append(out, lowerStmtAssign(st, v))
        case *ast.ReturnStmt:
            // Materialize return expressions so literals/ops appear as EXPR before RETURN
            var vals []ir.Value
            for i, e := range v.Results {
                if ex, ok := lowerExpr(st, e); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    // If returning Owned, wrap into handle (copy-on-new); prefer literal length when known
                    if st != nil && st.currentFn != "" {
                        if rts, ok2 := st.funcResults[st.currentFn]; ok2 && i < len(rts) {
                            rt := rts[i]
                            if rt == "Owned" || (len(rt) >= 6 && rt[:6] == "Owned<") {
                                // Determine data and length
                                var data ir.Value
                                var lenVal ir.Value
                                switch lit := e.(type) {
                                case *ast.StringLit:
                                    l := len(lit.Value)
                                    lenID := st.newTemp(); lres := &ir.Value{ID: lenID, Type: "int64"}
                                    out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", l), Result: lres})
                                    lenVal = *lres
                                    if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                                case *ast.SliceLit:
                                    l := len(lit.Elems)
                                    lenID := st.newTemp(); lres := &ir.Value{ID: lenID, Type: "int64"}
                                    out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", l), Result: lres})
                                    lenVal = *lres
                                    if ex.Result != nil { data = *ex.Result } else { data = ir.Value{ID: "", Type: "ptr"} }
                                default:
                                    if ex.Result != nil && (ex.Result.Type == "Owned" || (len(ex.Result.Type) >= 6 && ex.Result.Type[:6] == "Owned<")) {
                                        src := *ex.Result
                                        // ptr
                                        ptmp := st.newTemp(); pres := &ir.Value{ID: ptmp, Type: "ptr"}
                                        out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{src}, Result: pres})
                                        data = *pres
                                        // len
                                        ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                                        out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{src}, Result: lres})
                                        lenVal = *lres
                                    }
                                }
                                if lenVal.ID != "" {
                                    hid := st.newTemp(); hres := &ir.Value{ID: hid, Type: "Owned"}
                                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, lenVal}, Result: hres})
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
