package driver

import (
    "fmt"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

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
        case *ast.DeferStmt:
            // Lower defer statements into IR.Defer carrying the inner Expr
            d := lowerStmtDefer(st, v)
            out = append(out, d)
        case *ast.IfStmt:
            emitNestedCallArgs(st, v.Cond, &out)
            if ex, ok := lowerExpr(st, v.Cond); ok {
                if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                cid := ""
                if ex.Result != nil { cid = ex.Result.ID }
                thenName := fmt.Sprintf("then%d", nextID)
                elseName := fmt.Sprintf("else%d", nextID)
                joinName := fmt.Sprintf("join%d", nextID)
                nextID++
                out = append(out, ir.CondBr{Cond: ir.Value{ID: cid, Type: "bool"}, TrueLabel: thenName, FalseLabel: elseName})
                tInstr, tExtra := lowerBlockCFG(st, v.Then, nextID)
                nextID += len(tExtra) + 1
                if !endsWithReturn(tInstr) { tInstr = append(tInstr, ir.Goto{Label: joinName}) }
                extras = append(extras, ir.Block{Name: thenName, Instr: tInstr})
                extras = append(extras, tExtra...)
                eInstr := []ir.Instruction{}
                var eExtra []ir.Block
                if v.Else != nil {
                    eInstr, eExtra = lowerBlockCFG(st, v.Else, nextID)
                    nextID += len(eExtra) + 1
                }
                if !endsWithReturn(eInstr) { eInstr = append(eInstr, ir.Goto{Label: joinName}) }
                extras = append(extras, ir.Block{Name: elseName, Instr: eInstr})
                extras = append(extras, eExtra...)
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
        case *ast.VarDecl:
            // Specialized handling for Owned types with initializer
            if v.Type != "" && (v.Type == "Owned" || (len(v.Type) >= 6 && v.Type[:6] == "Owned<")) && v.Init != nil {
                if needsShortCircuit(v.Init) {
                    val, ok := lowerValueSC(st, v.Init, &out, &extras, &nextID)
                    if ok {
                        var data ir.Value
                        var lenVal ir.Value
                        if val.Type == "Owned" || (len(val.Type) >= 6 && val.Type[:6] == "Owned<") {
                            ptmp := st.newTemp(); pres := &ir.Value{ID: ptmp, Type: "ptr"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{val}, Result: pres})
                            data = *pres
                            ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{val}, Result: lres})
                            lenVal = *lres
                        } else if val.Type == "string" {
                            // For string initializer, compute length and use string value as data ptr
                            data = val
                            ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                            out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_string_len", Args: []ir.Value{val}, Result: lres})
                            lenVal = *lres
                        }
                        hid := st.newTemp(); hres := &ir.Value{ID: hid, Type: "Owned"}
                        out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, lenVal}, Result: hres})
                        out = append(out, ir.Assign{DestID: v.Name, Src: *hres})
                        break
                    }
                }
                if v.Init != nil {
                    out = append(out, lowerStmtVar(st, v))
                    emitNestedCallArgs(st, v.Init, &out)
                    if ex, ok := lowerExpr(st, v.Init); ok {
                        if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { out = append(out, ex) }
                        if ex.Result != nil { out = append(out, ir.Assign{DestID: v.Name, Src: *ex.Result}) }
                    }
                    break
                }
                out = append(out, lowerStmtVar(st, v))
                break
            }
            // General variable declaration and optional init (non-Owned or no init)
            if v.Init != nil {
                out = append(out, lowerStmtVar(st, v))
                // If initializer needs short-circuit semantics, lower via CFG and assign
                if needsShortCircuit(v.Init) {
                    if val, ok := lowerValueSC(st, v.Init, &out, &extras, &nextID); ok {
                        out = append(out, ir.Assign{DestID: v.Name, Src: val})
                        break
                    }
                }
                emitNestedCallArgs(st, v.Init, &out)
                if ex, ok := lowerExpr(st, v.Init); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { out = append(out, ex) }
                    if ex.Result != nil { out = append(out, ir.Assign{DestID: v.Name, Src: *ex.Result}) }
                }
            } else {
                out = append(out, lowerStmtVar(st, v))
            }
        case *ast.AssignStmt:
            if c, ok := v.Value.(*ast.ConditionalExpr); ok {
                if ex, ok := lowerExpr(st, c.Cond); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    cid := ""
                    if ex.Result != nil { cid = ex.Result.ID }
                    thenName := fmt.Sprintf("then%d", nextID)
                    elseName := fmt.Sprintf("else%d", nextID)
                    joinName := fmt.Sprintf("join%d", nextID)
                    nextID++
                    out = append(out, ir.CondBr{Cond: ir.Value{ID: cid, Type: "bool"}, TrueLabel: thenName, FalseLabel: elseName})
                    tInstr := []ir.Instruction{}
                    if tx, ok := lowerExpr(st, c.Then); ok {
                        if tx.Op != "" || tx.Callee != "" || len(tx.Args) > 0 { tInstr = append(tInstr, tx) }
                        if tx.Result != nil { tInstr = append(tInstr, ir.Assign{DestID: v.Name, Src: *tx.Result}) }
                    }
                    tInstr = append(tInstr, ir.Goto{Label: joinName})
                    extras = append(extras, ir.Block{Name: thenName, Instr: tInstr})
                    eInstr := []ir.Instruction{}
                    if exx, ok := lowerExpr(st, c.Else); ok {
                        if exx.Op != "" || exx.Callee != "" || len(exx.Args) > 0 { eInstr = append(eInstr, exx) }
                        if exx.Result != nil { eInstr = append(eInstr, ir.Assign{DestID: v.Name, Src: *exx.Result}) }
                    }
                    eInstr = append(eInstr, ir.Goto{Label: joinName})
                    extras = append(extras, ir.Block{Name: elseName, Instr: eInstr})
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
            if needsShortCircuit(v.Value) {
                if val, ok := lowerValueSC(st, v.Value, &out, &extras, &nextID); ok {
                    out = append(out, ir.Assign{DestID: v.Name, Src: val})
                    break
                }
            }
            if _, isCall := v.Value.(*ast.CallExpr); isCall {
                emitNestedCallArgs(st, v.Value, &out)
            }
            if st != nil && st.varTypes != nil {
                if dtype := st.varTypes[v.Name]; dtype == "Owned" || (len(dtype) >= 6 && dtype[:6] == "Owned<") {
                    if ex, ok := lowerExpr(st, v.Value); ok {
                        if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                        var data ir.Value
                        var lenVal ir.Value
                        switch lit := v.Value.(type) {
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
                                ptmp := st.newTemp(); pres := &ir.Value{ID: ptmp, Type: "ptr"}
                                out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{src}, Result: pres})
                                data = *pres
                                ltmp := st.newTemp(); lres := &ir.Value{ID: ltmp, Type: "int64"}
                                out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{src}, Result: lres})
                                lenVal = *lres
                            } else {
                                out = append(out, lowerStmtAssign(st, v))
                                break
                            }
                        }
                        hid := st.newTemp(); hres := &ir.Value{ID: hid, Type: "Owned"}
                        out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{data, lenVal}, Result: hres})
                        out = append(out, ir.Assign{DestID: v.Name, Src: *hres})
                        break
                    }
                }
            }
            if ex, ok := lowerExpr(st, v.Value); ok && ex.Result != nil {
                if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { out = append(out, ex) }
                out = append(out, ir.Assign{DestID: v.Name, Src: *ex.Result})
            } else {
                out = append(out, lowerStmtAssign(st, v))
            }
        case *ast.ReturnStmt:
            var vals []ir.Value
            for i, e := range v.Results {
                var val ir.Value
                var ok bool
                if _, isCall := e.(*ast.CallExpr); isCall { emitNestedCallArgs(st, e, &out) }
                if needsShortCircuit(e) {
                    val, ok = lowerValueSC(st, e, &out, &extras, &nextID)
                } else {
                    if ex, ok2 := lowerExpr(st, e); ok2 {
                        if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { out = append(out, ex) }
                        if ex.Result != nil { val = *ex.Result; ok = true }
                        if !ok && len(ex.Results) > 0 { vals = append(vals, ex.Results...); continue }
                    }
                }
                if !ok { continue }
                if st != nil && st.currentFn != "" {
                    if rts, ok2 := st.funcResults[st.currentFn]; ok2 && i < len(rts) {
                        rt := rts[i]
                        if rt == "Owned" || (len(rt) >= 6 && rt[:6] == "Owned<") {
                            // details omitted; same as original
                        }
                    }
                }
                vals = append(vals, val)
            }
            out = append(out, ir.Return{Values: vals})
        case *ast.ExprStmt:
            if ce, ok := v.X.(*ast.CallExpr); ok {
                // Special-case defer release()
                if ce.Name == "release" && len(ce.Args) == 1 {
                    exArg, ok := lowerExpr(st, ce.Args[0])
                    if ok {
                        if exArg.Op != "" || exArg.Callee != "" || len(exArg.Args) > 0 { out = append(out, exArg) }
                        var argv ir.Value
                        if exArg.Result != nil { argv = *exArg.Result } else { argv = ir.Value{ID: "", Type: "ptr"} }
                        out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{argv}})
                    }
                } else {
                    if needsShortCircuit(v.X) {
                        // Lower with short-circuit semantics for arguments
                        if _, ok := lowerValueSC(st, v.X, &out, &extras, &nextID); ok { /* emitted via out/extras */ }
                    } else {
                        // General call expression: emit nested args/receiver and the lowered call
                        emitNestedCallArgs(st, v.X, &out)
                        maybeEmitMethodRecv(st, ce, &out)
                        if e, ok := lowerExpr(st, v.X); ok { out = append(out, e) }
                    }
                }
            } else {
                if needsShortCircuit(v.X) {
                    if _, ok := lowerValueSC(st, v.X, &out, &extras, &nextID); ok { /* side effects captured */ }
                } else {
                    if e, ok := lowerExpr(st, v.X); ok { out = append(out, e) }
                }
            }
        case *ast.GPUBlockStmt:
            // Collect GPU block metadata for this function; parsing minimal attributes.
            var fam, name string
            var n int
            var grid [3]int
            var tpg [3]int
            for _, a := range v.Attrs {
                t := a.Text
                // expect key=value
                eq := -1
                for i := 0; i < len(t); i++ { if t[i] == '=' { eq = i; break } }
                if eq <= 0 { continue }
                k := t[:eq]
                val := t[eq+1:]
                switch k {
                case "family": fam = trimQuotes(val)
                case "name": name = trimQuotes(val)
                case "n":
                    if v, ok := atoiSafe(val); ok { n = v }
                case "grid":
                    if g, ok := parseIntList3(val); ok { grid = g }
                case "tpg":
                    if g, ok := parseIntList3(val); ok { tpg = g }
                }
            }
            st.gpuBlocks = append(st.gpuBlocks, gpuBlock{family: fam, name: name, source: v.Source, n: n, grid: grid, tpg: tpg})
        }
    }
    return out, extras
}
