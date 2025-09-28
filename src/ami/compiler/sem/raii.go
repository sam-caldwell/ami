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
        // Track releases by variable name → count and last position
        counts := map[string]int{}
        lastPos := map[string]diag.Position{}
        emitDouble := func(name string, pos diag.Position) {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "variable released more than once", Pos: &pos})
        }
        // Scan statements
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.DeferStmt:
                if name, p := releaseTargetFromCall(v.Call); name != "" {
                    counts[name]++
                    if counts[name] > 1 { emitDouble(name, diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}) }
                    lastPos[name] = diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}
                }
            case *ast.ExprStmt:
                if name, p := releaseTargetFromExpr(v.X); name != "" {
                    counts[name]++
                    if counts[name] > 1 { emitDouble(name, diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}) }
                    lastPos[name] = diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}
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
