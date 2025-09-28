package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
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
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        // Collect local variables/params for simple ownership presence checks.
        locals := map[string]bool{}
        for _, p := range fn.Params { if p.Name != "" { locals[p.Name] = true } }
        for _, st := range fn.Body.Stmts { if vd, ok := st.(*ast.VarDecl); ok && vd.Name != "" { locals[vd.Name] = true } }
        // Track releases by variable name → count and last position
        counts := map[string]int{}
        lastPos := map[string]diag.Position{}
        // Track variables released immediately; deferred releases do not immediately invalidate usage.
        releasedNow := map[string]bool{}
        emitDouble := func(name string, pos diag.Position) {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "variable released more than once", Pos: &pos})
        }
        // Scan statements
        for _, st := range fn.Body.Stmts {
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
                    if !locals[name] {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_RELEASE_UNOWNED", Message: "release of undeclared variable", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                    }
                    counts[name]++
                    if counts[name] > 1 { emitDouble(name, diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}) }
                    lastPos[name] = diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}
                    releasedNow[name] = true
                    continue
                }
                // detect use-after-release for immediate releases by scanning expression for idents
                if id := firstIdentUseInExpr(v.X, releasedNow); id != "" {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use after release: " + id, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
                }
            }
        }
    }
    return out
}

// releaseTargetFromExpr inspects an expression and returns the released variable name if the
// expression matches release(x) or mutate(release(x)). It also returns the position of the call.
func releaseTargetFromExpr(e ast.Expr) (string, source.Position) {
    switch v := e.(type) {
    case *ast.CallExpr:
        return releaseTargetFromCall(v)
    default:
        return "", source.Position{}
    }
}

// releaseTargetFromCall inspects a call expression for release(x) or mutate(release(x)).
func releaseTargetFromCall(c *ast.CallExpr) (string, source.Position) {
    if c == nil { return "", source.Position{} }
    if c.Name == "release" {
        if len(c.Args) >= 1 {
            if id, ok := c.Args[0].(*ast.IdentExpr); ok { return id.Name, c.NamePos }
        }
        return "", c.NamePos
    }
    if c.Name == "mutate" && len(c.Args) == 1 {
        if inner, ok := c.Args[0].(*ast.CallExpr); ok && inner.Name == "release" {
            if len(inner.Args) >= 1 {
                if id, ok := inner.Args[0].(*ast.IdentExpr); ok { return id.Name, c.NamePos }
            }
        }
    }
    return "", c.NamePos
}

// firstIdentUseInExpr returns the name of the first identifier used in expression `e`
// that appears in the `released` set. Returns empty string when none.
func firstIdentUseInExpr(e ast.Expr, released map[string]bool) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if released[v.Name] { return v.Name }
        return ""
    case *ast.CallExpr:
        for _, a := range v.Args { if n := firstIdentUseInExpr(a, released); n != "" { return n } }
        return ""
    case *ast.UnaryExpr:
        return firstIdentUseInExpr(v.X, released)
    case *ast.BinaryExpr:
        if n := firstIdentUseInExpr(v.X, released); n != "" { return n }
        return firstIdentUseInExpr(v.Y, released)
    default:
        return ""
    }
}
