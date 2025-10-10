package driver

import (
    "strings"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerInlineWorkers scans pipeline steps for inline func-literal workers and
// synthesizes named IR functions with deterministic names so that codegen can
// emit worker core wrappers. The generated body returns the input Event<T> and
// a nil error, which preserves types and enables execution via invoker path.
func lowerInlineWorkers(pkg, unit string, f *ast.File) []ir.Function {
    var out []ir.Function
    if f == nil { return out }
    // For each pipeline independently, count inline occurrences deterministically
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        inlineIdx := 0
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            lname := strings.ToLower(st.Name)
            if lname != "transform" && lname != "ingress" && lname != "egress" { continue }
            // find worker arg (named or first positional)
            var warg string
            has := false
            for _, a := range st.Args {
                if eq := strings.IndexByte(a.Text, '='); eq > 0 {
                    key := strings.TrimSpace(a.Text[:eq])
                    if strings.EqualFold(key, "worker") {
                        warg = strings.TrimSpace(a.Text[eq+1:])
                        has = true
                        break
                    }
                } else if !has {
                    warg = a.Text
                    has = true
                }
            }
            if !has { continue }
            if !strings.HasPrefix(strings.TrimSpace(warg), "func") { continue }
            // Parse signature minimally from the literal text
            warg = strings.TrimSpace(warg)
            paramTyp, results, ok := semInlineSig(warg)
            if !ok { continue }
            tParam := extractEventArg(paramTyp)
            if tParam == "" { tParam = "any" }
            // Determine result types: default to (Event<T>, error) but honor explicit results
            var r0Ty, r1Ty string
            if len(results) > 0 {
                r0Ty = strings.TrimSpace(results[0])
                if len(results) >= 2 {
                    r1Ty = strings.TrimSpace(results[1])
                }
            }
            if r0Ty == "" { r0Ty = "Event<" + tParam + ">" }
            if r1Ty == "" { r1Ty = "error" }
            // Deterministic name consistent with pipelines debug normalization
            inlineIdx++
            name := "InlineWorker_" + unit + "_" + pd.Name + "_" + itoa(inlineIdx)
            // Build IR function: func(ev Event<T>) (R0, error) with body lowered from inline literal (safe subset)
            evTy := "Event<" + tParam + ">"
            fn := ir.Function{Name: name}
            fn.Params = []ir.Value{{ID: "ev", Type: evTy}}
            fn.Results = []ir.Value{{ID: "r0", Type: r0Ty}, {ID: "r1", Type: r1Ty}}
            // Try to extract a return expression or if/else from the literal body and lower to IR
            body := extractInlineBody(warg)
            var instr []ir.Instruction
            // Always begin by producing a null error into err0
            errID := "err0"
            instr = append(instr, ir.Expr{Op: "field.null", Result: &ir.Value{ID: errID, Type: "error"}})
            // Lower minimal forms:
            // - return ev[, nil]
            // - return <int|real|bool|string-literal>[, nil]
            // - return <lit> <op> <lit>[, nil] where op in +,-,*,/,% and result is numeric
            lowered := false
            // If/else branch form
            if br, ok := parseInlineIfReturn(body); ok {
                // Only support simple condition on numeric literals and branch returns of ev or literal matching r0Ty
                // Build entry: compute condition
                condTy := "bool"
                // Use int or real literal type depending on dots
                numTy := "int"
                if strings.Contains(br.lhs, ".") || strings.Contains(br.rhs, ".") { numTy = "real" }
                // entry block
                var entry ir.Block
                // err per-branch; not needed in entry
                // build literals and comparison
                lhsID, rhsID, cID := "c0", "c1", "cond"
                entry.Instr = append(entry.Instr, ir.Expr{Op: "lit:" + br.lhs, Result: &ir.Value{ID: lhsID, Type: numTy}})
                entry.Instr = append(entry.Instr, ir.Expr{Op: "lit:" + br.rhs, Result: &ir.Value{ID: rhsID, Type: numTy}})
                cop := map[string]string{"==": "eq", "!=": "ne", "<": "lt", "<=": "le", ">": "gt", ">=": "ge"}[br.op]
                entry.Instr = append(entry.Instr, ir.Expr{Op: cop, Args: []ir.Value{{ID: lhsID, Type: numTy}, {ID: rhsID, Type: numTy}}, Result: &ir.Value{ID: cID, Type: condTy}})
                entry.Instr = append(entry.Instr, ir.CondBr{Cond: ir.Value{ID: cID, Type: condTy}, TrueLabel: "then", FalseLabel: "else"})
                // then block
                var then ir.Block
                then.Name = "then"
                // err0 in then
                then.Instr = append(then.Instr, ir.Expr{Op: "field.null", Result: &ir.Value{ID: errID, Type: "error"}})
                if br.thenIsEv && strings.HasPrefix(r0Ty, "Event<") {
                    then.Instr = append(then.Instr, ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}})
                } else if !br.thenIsEv && isPrimitiveLike(r0Ty) {
                    vID := "t0"
                    then.Instr = append(then.Instr, ir.Expr{Op: "lit:" + br.thenLit, Result: &ir.Value{ID: vID, Type: r0Ty}})
                    then.Instr = append(then.Instr, ir.Return{Values: []ir.Value{{ID: vID, Type: r0Ty}, {ID: errID, Type: "error"}}})
                } else {
                    // fallback: identity event
                    then.Instr = append(then.Instr, ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}})
                }
                // else block
                var els ir.Block
                els.Name = "else"
                els.Instr = append(els.Instr, ir.Expr{Op: "field.null", Result: &ir.Value{ID: errID, Type: "error"}})
                if br.elseIsEv && strings.HasPrefix(r0Ty, "Event<") {
                    els.Instr = append(els.Instr, ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}})
                } else if !br.elseIsEv && isPrimitiveLike(r0Ty) {
                    vID := "e0"
                    els.Instr = append(els.Instr, ir.Expr{Op: "lit:" + br.elseLit, Result: &ir.Value{ID: vID, Type: r0Ty}})
                    els.Instr = append(els.Instr, ir.Return{Values: []ir.Value{{ID: vID, Type: r0Ty}, {ID: errID, Type: "error"}}})
                } else {
                    els.Instr = append(els.Instr, ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}})
                }
                fn.Blocks = []ir.Block{entry, then, els}
                lowered = true
            } else if r, ok := parseInlineReturn(body); ok {
                switch r.kind {
                case retEV:
                    // Only valid if result0 expects an Event
                    if strings.HasPrefix(r0Ty, "Event<") {
                        fn.Blocks = []ir.Block{{Name: "entry", Instr: append(instr, ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}})}}
                        lowered = true
                    }
                case retLit:
                    // Literal return; honor r0Ty when primitive-like, else fall back
                    if isPrimitiveLike(r0Ty) {
                        valID := "v0"
                        instr = append(instr, ir.Expr{Op: "lit:" + r.lit, Result: &ir.Value{ID: valID, Type: r0Ty}})
                        fn.Blocks = []ir.Block{{Name: "entry", Instr: append(instr, ir.Return{Values: []ir.Value{{ID: valID, Type: r0Ty}, {ID: errID, Type: "error"}}})}}
                        lowered = true
                    }
                case retBinOp:
                    if isNumericLike(r0Ty) {
                        // Support LHS/RHS from literals or Event payload (when marked)
                        lhsID := "c0"
                        rhsID := "c1"
                        resID := "v0"
                        if r.lhsIsEv {
                            // extract numeric payload from ev
                            pid := "p0"
                            instr = append(instr, ir.Expr{Op: "event.payload", Args: []ir.Value{{ID: "ev", Type: evTy}}, Result: &ir.Value{ID: pid, Type: r0Ty}})
                            lhsID = pid
                        } else {
                            instr = append(instr, ir.Expr{Op: "lit:" + r.lhs, Result: &ir.Value{ID: lhsID, Type: r0Ty}})
                        }
                        if r.rhsIsEv {
                            pid := "p1"
                            instr = append(instr, ir.Expr{Op: "event.payload", Args: []ir.Value{{ID: "ev", Type: evTy}}, Result: &ir.Value{ID: pid, Type: r0Ty}})
                            rhsID = pid
                        } else {
                            instr = append(instr, ir.Expr{Op: "lit:" + r.rhs, Result: &ir.Value{ID: rhsID, Type: r0Ty}})
                        }
                        op := map[string]string{"+": "add", "-": "sub", "*": "mul", "/": "div", "%": "mod"}[r.op]
                        instr = append(instr, ir.Expr{Op: op, Args: []ir.Value{{ID: lhsID, Type: r0Ty}, {ID: rhsID, Type: r0Ty}}, Result: &ir.Value{ID: resID, Type: r0Ty}})
                        fn.Blocks = []ir.Block{{Name: "entry", Instr: append(instr, ir.Return{Values: []ir.Value{{ID: resID, Type: r0Ty}, {ID: errID, Type: "error"}}})}}
                        lowered = true
                    }
                case retCmp:
                    // Comparisons yield boolean; when r0Ty not explicitly bool, coerce to bool
                    boolTy := r0Ty
                    if strings.TrimSpace(boolTy) == "" || !strings.EqualFold(boolTy, "bool") {
                        boolTy = "bool"
                    }
                    // Use numeric type determination: if lhs contains '.', assume real; else int
                    numTy := "int"
                    if strings.Contains(r.lhs, ".") || strings.Contains(r.rhs, ".") { numTy = "real" }
                    lhsID := "c0"
                    rhsID := "c1"
                    resID := "v0"
                    instr = append(instr, ir.Expr{Op: "lit:" + r.lhs, Result: &ir.Value{ID: lhsID, Type: numTy}})
                    instr = append(instr, ir.Expr{Op: "lit:" + r.rhs, Result: &ir.Value{ID: rhsID, Type: numTy}})
                    op := map[string]string{"==": "eq", "!=": "ne", "<": "lt", "<=": "le", ">": "gt", ">=": "ge"}[r.op]
                    instr = append(instr, ir.Expr{Op: op, Args: []ir.Value{{ID: lhsID, Type: numTy}, {ID: rhsID, Type: numTy}}, Result: &ir.Value{ID: resID, Type: boolTy}})
                    fn.Blocks = []ir.Block{{Name: "entry", Instr: append(instr, ir.Return{Values: []ir.Value{{ID: resID, Type: boolTy}, {ID: errID, Type: "error"}}})}}
                    lowered = true
                }
            }
            if !lowered {
                // Fallback: identity Event return to preserve previous behavior
                // If r0Ty is not an Event, produce a zero literal for primitive-like types.
                if strings.HasPrefix(r0Ty, "Event<") {
                    fn.Blocks = []ir.Block{{Name: "entry", Instr: append(instr, ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}})}}
                } else if isPrimitiveLike(r0Ty) {
                    valID := "v0"
                    lit := "0"
                    if r0Ty == "real" || r0Ty == "float64" { lit = "0.0" }
                    instr = append(instr, ir.Expr{Op: "lit:" + lit, Result: &ir.Value{ID: valID, Type: r0Ty}})
                    fn.Blocks = []ir.Block{{Name: "entry", Instr: append(instr, ir.Return{Values: []ir.Value{{ID: valID, Type: r0Ty}, {ID: errID, Type: "error"}}})}}
                } else {
                    fn.Blocks = []ir.Block{{Name: "entry", Instr: append(instr, ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}})}}
                }
            }
            out = append(out, fn)
        }
    }
    return out
}
