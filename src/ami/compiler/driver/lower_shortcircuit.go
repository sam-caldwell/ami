package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "fmt"
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

// lowerValueSC lowers e into instructions with short-circuit semantics and returns
// a Value holding the result. Instructions are appended to out/extras. nextID seeds labels.
func lowerValueSC(st *lowerState, e ast.Expr, out *[]ir.Instruction, extras *[]ir.Block, nextID *int) (ir.Value, bool) {
    switch v := e.(type) {
    case *ast.ConditionalExpr:
        // cond
        cx, ok := lowerExpr(st, v.Cond)
        if !ok { return ir.Value{}, false }
        if cx.Op != "" || cx.Callee != "" || len(cx.Args) > 0 { *out = append(*out, cx) }
        cid := ""
        if cx.Result != nil { cid = cx.Result.ID }
        thenName := fmt.Sprintf("then%d", *nextID)
        elseName := fmt.Sprintf("else%d", *nextID)
        joinName := fmt.Sprintf("join%d", *nextID)
        *nextID++
        *out = append(*out, ir.CondBr{Cond: ir.Value{ID: cid, Type: "bool"}, TrueLabel: thenName, FalseLabel: elseName})
        // then
        tInstr := []ir.Instruction{}
        tv, ok := lowerValueSC(st, v.Then, &tInstr, extras, nextID)
        if !ok { return ir.Value{}, false }
        tInstr = append(tInstr, ir.Goto{Label: joinName})
        *extras = append(*extras, ir.Block{Name: thenName, Instr: tInstr})
        // else
        eInstr := []ir.Instruction{}
        ev, ok := lowerValueSC(st, v.Else, &eInstr, extras, nextID)
        if !ok { return ir.Value{}, false }
        eInstr = append(eInstr, ir.Goto{Label: joinName})
        *extras = append(*extras, ir.Block{Name: elseName, Instr: eInstr})
        // join with phi
        rid := st.newTemp()
        rtype := tv.Type
        if rtype == "" || (ev.Type != "" && ev.Type != rtype) { rtype = "any" }
        r := ir.Value{ID: rid, Type: rtype}
        phi := ir.Phi{Result: r, Incomings: []ir.PhiIncoming{{Value: tv, Label: thenName}, {Value: ev, Label: elseName}}}
        *extras = append(*extras, ir.Block{Name: joinName, Instr: []ir.Instruction{phi}})
        return r, true
    case *ast.BinaryExpr:
        // boolean short-circuit
        if v.Op == token.And || v.Op == token.Or {
            // left value
            lx, ok := lowerValueSC(st, v.X, out, extras, nextID)
            if !ok { return ir.Value{}, false }
            // branch on left
            thenName := fmt.Sprintf("sc_then%d", *nextID)
            elseName := fmt.Sprintf("sc_else%d", *nextID)
            joinName := fmt.Sprintf("sc_join%d", *nextID)
            *nextID++
            // For &&: true → eval RHS; false → 0
            // For ||: true → 1; false → eval RHS
            *out = append(*out, ir.CondBr{Cond: lx, TrueLabel: thenName, FalseLabel: elseName})
            // Build branches
            var tInstr, eInstr []ir.Instruction
            var tv, ev ir.Value
            if v.Op == token.And {
                // then: eval RHS
                tv, ok = lowerValueSC(st, v.Y, &tInstr, extras, nextID)
                if !ok { return ir.Value{}, false }
                // else: false
                fid := st.newTemp(); fres := &ir.Value{ID: fid, Type: "bool"}
                eInstr = append(eInstr, ir.Expr{Op: "lit:0", Result: fres})
                ev = *fres
            } else { // Or
                // then: true
                tid := st.newTemp(); tres := &ir.Value{ID: tid, Type: "bool"}
                tInstr = append(tInstr, ir.Expr{Op: "lit:1", Result: tres})
                tv = *tres
                // else: eval RHS
                ev, ok = lowerValueSC(st, v.Y, &eInstr, extras, nextID)
                if !ok { return ir.Value{}, false }
            }
            tInstr = append(tInstr, ir.Goto{Label: joinName})
            eInstr = append(eInstr, ir.Goto{Label: joinName})
            *extras = append(*extras, ir.Block{Name: thenName, Instr: tInstr})
            *extras = append(*extras, ir.Block{Name: elseName, Instr: eInstr})
            rid := st.newTemp()
            r := ir.Value{ID: rid, Type: "bool"}
            phi := ir.Phi{Result: r, Incomings: []ir.PhiIncoming{{Value: tv, Label: thenName}, {Value: ev, Label: elseName}}}
            *extras = append(*extras, ir.Block{Name: joinName, Instr: []ir.Instruction{phi}})
            return r, true
        }
        // fallback to eager lowering for non-boolean binary ops
        ex, ok := lowerExpr(st, v)
        if !ok || ex.Result == nil { return ir.Value{}, false }
        *out = append(*out, ex)
        return *ex.Result, true
    case *ast.CallExpr:
        // Lower args with short-circuit as needed
        var args []ir.Value
        for _, a := range v.Args {
            if av, ok := lowerValueSC(st, a, out, extras, nextID); ok {
                args = append(args, av)
            }
        }
        // Determine result type from signatures if available
        var res *ir.Value
        if st != nil && st.funcResults != nil {
            if rs, ok := st.funcResults[v.Name]; ok && len(rs) > 0 && rs[0] != "" {
                id := st.newTemp()
                res = &ir.Value{ID: id, Type: rs[0]}
            } else {
                res = nil // void call
            }
        }
        *out = append(*out, ir.Expr{Op: "call", Callee: v.Name, Args: args, Result: res})
        if res != nil { return *res, true }
        return ir.Value{}, true
    default:
        // generic eager lowering; safe for literals/idents/unary
        ex, ok := lowerExpr(st, e)
        if !ok || ex.Result == nil { return ir.Value{}, false }
        *out = append(*out, ex)
        return *ex.Result, true
    }
}
