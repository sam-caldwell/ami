package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "fmt"
)

// lowerBlock lowers a function body block into a sequence of IR instructions.
func lowerBlock(st *lowerState, b *ast.BlockStmt) []ir.Instruction {
    instrs, _ := lowerBlockCFG(st, b, 0)
    return instrs
}

// lowerBlockCFG lowers a block into entry instructions plus any extra blocks for control flow.
// It returns the entry instructions and a slice of additional blocks that callers can append
// to the function. The blockId seeds unique label names.
func lowerBlockCFG(st *lowerState, b *ast.BlockStmt, blockId int) ([]ir.Instruction, []ir.Block) {
    var out []ir.Instruction
    var extras []ir.Block
    if b == nil { return out, extras }
    nextID := blockId
    endsWithReturn := func(ins []ir.Instruction) bool {
        if len(ins) == 0 { return false }
        _, ok := ins[len(ins)-1].(ir.Return)
        return ok
    }
    for i := 0; i < len(b.Stmts); i++ {
        s := b.Stmts[i]
        switch v := s.(type) {
        case *ast.IfStmt:
            // Lower condition
            // Emit any nested call expressions used as arguments inside the condition before the condition itself
            emitNestedCallArgs(st, v.Cond, &out)
            if ex, ok := lowerExpr(st, v.Cond); ok {
                if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                cid := ""
                if ex.Result != nil { cid = ex.Result.ID }
                // Create labels
                thenName := fmt.Sprintf("then%d", nextID)
                elseName := fmt.Sprintf("else%d", nextID)
                joinName := fmt.Sprintf("join%d", nextID)
                nextID++
                // Branch
                out = append(out, ir.CondBr{Cond: ir.Value{ID: cid, Type: "bool"}, TrueLabel: thenName, FalseLabel: elseName})
                // Lower then block
                tInstr, tExtra := lowerBlockCFG(st, v.Then, nextID)
                nextID += len(tExtra) + 1
                if !endsWithReturn(tInstr) { tInstr = append(tInstr, ir.Goto{Label: joinName}) }
                extras = append(extras, ir.Block{Name: thenName, Instr: tInstr})
                extras = append(extras, tExtra...)
                // Lower else block
                eInstr := []ir.Instruction{}
                var eExtra []ir.Block
                if v.Else != nil {
                    eInstr, eExtra = lowerBlockCFG(st, v.Else, nextID)
                    nextID += len(eExtra) + 1
                }
                if !endsWithReturn(eInstr) { eInstr = append(eInstr, ir.Goto{Label: joinName}) }
                extras = append(extras, ir.Block{Name: elseName, Instr: eInstr})
                extras = append(extras, eExtra...)
                // Lower remaining statements in this block into the join block
                var rest *ast.BlockStmt
                if i+1 < len(b.Stmts) {
                    rest = &ast.BlockStmt{Stmts: b.Stmts[i+1:]}
                }
                joinInstr := []ir.Instruction{}
                var joinExtra []ir.Block
                if rest != nil {
                    joinInstr, joinExtra = lowerBlockCFG(st, rest, nextID)
                    nextID += len(joinExtra) + 1
                }
                extras = append(extras, ir.Block{Name: joinName, Instr: joinInstr})
                extras = append(extras, joinExtra...)
                // Entire remainder handled via join; terminate processing of this lexical block
                return out, extras
            }

        case *ast.VarDecl:
            // Owned ABI: unconditional copy-on-new for Owned variables with initializers.
            if v.Type != "" && (v.Type == "Owned" || (len(v.Type) >= 6 && v.Type[:6] == "Owned<")) && v.Init != nil {
                // Short-circuit aware lowering for initializer
                if needsShortCircuit(v.Init) {
                    // produce the chosen value
                    val, ok := lowerValueSC(st, v.Init, &out, &extras, &nextID)
                    if ok {
                        var data ir.Value
                        var lenVal ir.Value
                        // Determine length depending on source kind
                        // If existing Owned handle, derive via runtime helpers
                        if val.Type == "Owned" || (len(val.Type) >= 6 && val.Type[:6] == "Owned<") {
                            // ptr
                            ptmp := st.newTemp(); pres := &ir.Value{ID: ptmp, Type: "ptr"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{val}, Result: pres})
                            data = *pres
                            // len
                            ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{val}, Result: lres})
                            lenVal = *lres
                        } else if val.Type == "string" {
                            data = val
                            ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_string_len", Args: []ir.Value{val}, Result: lres})
                            lenVal = *lres
                        } else if len(val.Type) >= 6 && val.Type[:6] == "slice<" { // slice<T>
                            data = val
                            ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_slice_len", Args: []ir.Value{val}, Result: lres})
                            lenVal = *lres
                        }
                        if lenVal.ID != "" {
                            hid := st.newTemp(); hres := &ir.Value{ID: hid, Type: "Owned"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, lenVal}, Result: hres})
                            // emit var using the owned handle
                            res := ir.Value{ID: v.Name, Type: v.Type}
                            out = append(out, ir.Var{Name: v.Name, Type: v.Type, Init: hres, Result: res})
                            if st != nil && st.varTypes != nil && v.Name != "" { st.varTypes[v.Name] = v.Type }
                            break
                        }
                    }
                } else if ex, ok := lowerExpr(st, v.Init); ok {
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
                // Non-Owned var: allow short-circuit initializer
                if v.Init != nil && needsShortCircuit(v.Init) {
                    if val, ok := lowerValueSC(st, v.Init, &out, &extras, &nextID); ok {
                        vtype := v.Type
                        if vtype == "" { vtype = val.Type }
                        res := ir.Value{ID: v.Name, Type: vtype}
                        out = append(out, ir.Var{Name: v.Name, Type: vtype, Init: &val, Result: res})
                        if st != nil && st.varTypes != nil && v.Name != "" && vtype != "" { st.varTypes[v.Name] = vtype }
                        break
                    }
                }
                // General var with initializer: ensure initializer expression and any nested call-arguments are emitted.
                if v.Init != nil {
                    // First, declare the var so it exists for subsequent assign
                    out = append(out, lowerStmtVar(st, v))
                    // Emit nested call arguments (one or more levels) before lowering the initializer itself
                    emitNestedCallArgs(st, v.Init, &out)
                    // Now lower the initializer and assign to the variable
                    if ex, ok := lowerExpr(st, v.Init); ok {
                        if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { out = append(out, ex) }
                        if ex.Result != nil { out = append(out, ir.Assign{DestID: v.Name, Src: *ex.Result}) }
                    }
                    break
                }
                // No initializer: just declare the var
                out = append(out, lowerStmtVar(st, v))
            }
        case *ast.AssignStmt:
            // Ternary conditional assignment: x = (cond ? a : b)
            if c, ok := v.Value.(*ast.ConditionalExpr); ok {
                // Lower condition first and branch to then/else blocks, assigning in each.
                if ex, ok := lowerExpr(st, c.Cond); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    cid := ""
                    if ex.Result != nil { cid = ex.Result.ID }
                    thenName := fmt.Sprintf("then%d", nextID)
                    elseName := fmt.Sprintf("else%d", nextID)
                    joinName := fmt.Sprintf("join%d", nextID)
                    nextID++
                    out = append(out, ir.CondBr{Cond: ir.Value{ID: cid, Type: "bool"}, TrueLabel: thenName, FalseLabel: elseName})
                    // then block: lower then-expr and assign
                    tInstr := []ir.Instruction{}
                    if tx, ok := lowerExpr(st, c.Then); ok {
                        if tx.Op != "" || tx.Callee != "" || len(tx.Args) > 0 { tInstr = append(tInstr, tx) }
                        if tx.Result != nil { tInstr = append(tInstr, ir.Assign{DestID: v.Name, Src: *tx.Result}) }
                    }
                    tInstr = append(tInstr, ir.Goto{Label: joinName})
                    extras = append(extras, ir.Block{Name: thenName, Instr: tInstr})
                    // else block
                    eInstr := []ir.Instruction{}
                    if exx, ok := lowerExpr(st, c.Else); ok {
                        if exx.Op != "" || exx.Callee != "" || len(exx.Args) > 0 { eInstr = append(eInstr, exx) }
                        if exx.Result != nil { eInstr = append(eInstr, ir.Assign{DestID: v.Name, Src: *exx.Result}) }
                    }
                    eInstr = append(eInstr, ir.Goto{Label: joinName})
                    extras = append(extras, ir.Block{Name: elseName, Instr: eInstr})
                    // Lower the remainder of statements into the join block
                    var rest *ast.BlockStmt
                    if i+1 < len(b.Stmts) { rest = &ast.BlockStmt{Stmts: b.Stmts[i+1:]} }
                    joinInstr := []ir.Instruction{}
                    var joinExtra []ir.Block
                    if rest != nil {
                        joinInstr, joinExtra = lowerBlockCFG(st, rest, nextID)
                        nextID += len(joinExtra) + 1
                    }
                    extras = append(extras, ir.Block{Name: joinName, Instr: joinInstr})
                    extras = append(extras, joinExtra...)
                    return out, extras
                }
            }
            // Generic short-circuit evaluation on RHS for assignments when not Owned-specialized
            if needsShortCircuit(v.Value) {
                if val, ok := lowerValueSC(st, v.Value, &out, &extras, &nextID); ok {
                    out = append(out, ir.Assign{DestID: v.Name, Src: val})
                    break
                }
            }
            // When RHS is a call, emit any nested call arguments before lowering RHS
            if _, isCall := v.Value.(*ast.CallExpr); isCall {
                emitNestedCallArgs(st, v.Value, &out)
            }
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
            // Lower RHS and then assign
            if ex, ok := lowerExpr(st, v.Value); ok && ex.Result != nil {
                if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { out = append(out, ex) }
                out = append(out, ir.Assign{DestID: v.Name, Src: *ex.Result})
            } else {
                out = append(out, lowerStmtAssign(st, v))
            }
        case *ast.ReturnStmt:
            // Materialize return expressions with short-circuit semantics when needed.
            var vals []ir.Value
            for i, e := range v.Results {
                var val ir.Value
                var ok bool
                // If returning a call expression, ensure its nested call-arguments are emitted first
                if _, isCall := e.(*ast.CallExpr); isCall {
                    emitNestedCallArgs(st, e, &out)
                }
                if needsShortCircuit(e) {
                    val, ok = lowerValueSC(st, e, &out, &extras, &nextID)
                } else {
                    if ex, ok2 := lowerExpr(st, e); ok2 {
                        if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { out = append(out, ex) }
                        if ex.Result != nil { val = *ex.Result; ok = true }
                        if !ok && len(ex.Results) > 0 {
                            // When call returns multiple values, append all to return vector directly
                            vals = append(vals, ex.Results...)
                            continue
                        }
                    }
                }
                if !ok { continue }
                // Owned return copy-on-new
                if st != nil && st.currentFn != "" {
                    if rts, ok2 := st.funcResults[st.currentFn]; ok2 && i < len(rts) {
                        rt := rts[i]
                        if rt == "Owned" || (len(rt) >= 6 && rt[:6] == "Owned<") {
                            var data ir.Value
                            var lenVal ir.Value
                            switch lit := e.(type) {
                            case *ast.StringLit:
                                l := len(lit.Value)
                                lenID := st.newTemp(); lres := &ir.Value{ID: lenID, Type: "int64"}
                                out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", l), Result: lres})
                                lenVal = *lres
                                data = val
                            case *ast.SliceLit:
                                l := len(lit.Elems)
                                lenID := st.newTemp(); lres := &ir.Value{ID: lenID, Type: "int64"}
                                out = append(out, ir.Expr{Op: fmt.Sprintf("lit:%d", l), Result: lres})
                                lenVal = *lres
                                data = val
                            default:
                                // derive ptr/len when returning an Owned handle
                                if val.Type == "Owned" || (len(val.Type) >= 6 && val.Type[:6] == "Owned<") {
                                    // ptr
                                    ptmp := st.newTemp(); pres := &ir.Value{ID: ptmp, Type: "ptr"}
                                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{val}, Result: pres})
                                    data = *pres
                                    // len
                                    ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{val}, Result: lres})
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
                vals = append(vals, val)
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
                if needsShortCircuit(v.X) {
                    // Compute value for its side effect (calls) with short-circuit args
                    if _, ok := lowerValueSC(st, v.X, &out, &extras, &nextID); ok { /* done */ }
                } else {
                    // Emit nested call arguments (if any) before lowering the expression itself
                    if _, isCall := v.X.(*ast.CallExpr); isCall {
                        emitNestedCallArgs(st, v.X, &out)
                    }
                    if e, ok := lowerExpr(st, v.X); ok { out = append(out, e) }
                }
            }
        }
    }
    return out, extras
}

// emitNestedCallArgs lowers and emits any nested call expressions used as arguments
// within the provided expression, ensuring they appear before the outer expression.
// This is needed so that argument temporaries are defined before use when lowering
// function calls like f(g(x), h(y)).
func emitNestedCallArgs(st *lowerState, e ast.Expr, out *[]ir.Instruction) {
    // Recurse through call arguments; do not emit the top-level call here.
    var walkArgs func(ast.Expr)
    walkArgs = func(x ast.Expr) {
        if ce, ok := x.(*ast.CallExpr); ok {
            // First, descend into grandchildren
            for _, a := range ce.Args { walkArgs(a) }
            // Then, emit this call so its result is available for parent
            if ex, ok2 := lowerExpr(st, ce); ok2 {
                if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 {
                    *out = append(*out, ex)
                }
            }
            return
        }
        // For non-call nodes that contain expressions (e.g., selector or binary),
        // add cases here as language grows. Currently only calls need eager emission.
    }
    // Start by locating the immediate call (if any) and walking its args
    if ce, ok := e.(*ast.CallExpr); ok {
        for _, a := range ce.Args { walkArgs(a) }
    }
}
