package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeRAII scans function bodies for release semantics with defer scheduling.
// It recognizes release(x) and mutate(release(x)) in both immediate and deferred
// forms. It emits E_RAII_DOUBLE_RELEASE if the same variable is released more
// than once (including combinations of immediate and deferred releases).
//
// Notes:
// - Missing-release detection for Owned<T> is deferred until generic var typing
//   is fully captured by the parser. This pass focuses on double-release safety
//   and correct accounting of deferred releases.
func AnalyzeRAII(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // Collect package‑local function signatures to identify Owned results/params for transfer/leak accounting.
    results := collectFunctionResults(f)
    params := collectFunctionParams(f)
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        // Collect local variables/params for simple ownership presence checks.
        locals := map[string]bool{}
        owned := map[string]bool{}       // variables considered Owned-managed
        consumed := map[string]bool{}    // variables released or transferred
        transferred := map[string]bool{} // variables transferred out (moved)
        varPos := map[string]diag.Position{}
        for _, p := range fn.Params {
            if p.Name != "" {
                locals[p.Name] = true
                varPos[p.Name] = diag.Position{Line: p.Pos.Line, Column: p.Pos.Column, Offset: p.Pos.Offset}
                // Treat textual type "Owned" as Owned-managed
                if p.Type == "Owned" { owned[p.Name] = true }
            }
        }
        for _, st := range fn.Body.Stmts {
            if vd, ok := st.(*ast.VarDecl); ok && vd.Name != "" {
                locals[vd.Name] = true
                varPos[vd.Name] = diag.Position{Line: vd.Pos.Line, Column: vd.Pos.Column, Offset: vd.Pos.Offset}
                // Explicit Owned type (without generics) counts as Owned-managed
                if vd.Type == "Owned" { owned[vd.Name] = true }
            }
        }
        // Track releases by variable name → count and last position
        counts := map[string]int{}
        lastPos := map[string]diag.Position{}
        // Track variables released immediately; deferred releases do not immediately invalidate usage.
        releasedNow := map[string]bool{}
        emitDouble := func(name string, pos diag.Position) {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "variable released more than once", Pos: &pos})
        }
        // helper: mark transfer when call param at argPos expects Owned
        transferOwned := func(call *ast.CallExpr, arg ast.Expr, argPos int) {
            if call == nil { return }
            sigp := params[call.Name]
            if argPos >= 0 && argPos < len(sigp) && sigp[argPos] == "Owned" {
                if id, ok := arg.(*ast.IdentExpr); ok {
                    name := id.Name
                    if !owned[name] {
                        // transferring an unowned variable
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_TRANSFER_UNOWNED", Message: "transfer of unowned variable", Pos: &diag.Position{Line: id.Pos.Line, Column: id.Pos.Column, Offset: id.Pos.Offset}})
                    }
                    consumed[name] = true
                    transferred[name] = true
                }
            }
        }
        // Scan statements
        for _, st := range fn.Body.Stmts {
            // Strict use-after-transfer: if any transferred variable appears in this statement, flag it.
            if name, pos := firstTransferredUseInStmt(st, transferred); name != "" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_USE_AFTER_TRANSFER", Message: "use after transfer: " + name, Pos: &diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset}})
            }
            switch v := st.(type) {
            case *ast.DeferStmt:
                if name, p := releaseTargetFromCall(v.Call); name != "" {
                    // simple guard: releasing unknown local
                    if !locals[name] {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_RELEASE_UNOWNED", Message: "release of undeclared variable", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                    }
                    counts[name]++
                    if counts[name] > 1 { emitDouble(name, diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}) }
                    lastPos[name] = diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}
                }
            case *ast.ExprStmt:
                if name, p := releaseTargetFromExpr(v.X); name != "" {
                    if transferred[name] {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_RELEASE_AFTER_TRANSFER", Message: "release after transfer: " + name, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                    }
                    if !locals[name] {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_RELEASE_UNOWNED", Message: "release of undeclared variable", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                    }
                    counts[name]++
                    if counts[name] > 1 { emitDouble(name, diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}) }
                    lastPos[name] = diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}
                    releasedNow[name] = true
                    consumed[name] = true
                    continue
                }
                // identify transfers for calls
                if ce, ok := v.X.(*ast.CallExpr); ok {
                    for i, a := range ce.Args { transferOwned(ce, a, i) }
                }
                // detect use-after-release for immediate releases by scanning expression for idents
                if id := firstIdentUseInExpr(v.X, releasedNow); id != "" {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use after release: " + id, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
                }
            case *ast.AssignStmt:
                // Acquisition: if RHS is call whose first result type is Owned, mark LHS as owned
                if ce, ok := v.Value.(*ast.CallExpr); ok {
                    if rs, ok := results[ce.Name]; ok && len(rs) > 0 && rs[0] == "Owned" {
                        owned[v.Name] = true
                        if _, exists := varPos[v.Name]; !exists { varPos[v.Name] = diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset} }
                    }
                    // Also consider transfer through arguments
                    for i, a := range ce.Args { transferOwned(ce, a, i) }
                }
                // Strict: assignment to a transferred variable is disallowed
                if transferred[v.Name] {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_ASSIGN_AFTER_TRANSFER", Message: "assignment to variable after transfer: " + v.Name, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
                }
            }
        }
        // Leak detection: any variable marked owned but not consumed (released or transferred)
        for name, isOwned := range owned {
            if !isOwned { continue }
            if consumed[name] { continue }
            // Emit leak at declaration/param position or function name when unknown
            p := varPos[name]
            if p.Line == 0 { p = diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset} }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_LEAK", Message: "Owned variable not released or transferred: " + name, Pos: &p})
        }
    }
    return out
}

// releaseTargetFromExpr inspects an expression and returns the released variable name if the
// expression matches release(x) or mutate(release(x)). It also returns the position of the call.
