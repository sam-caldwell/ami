package driver

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// writeIRRAIIDebug emits a per-unit RAII trace file under build/debug/ir/<pkg>/<unit>.raii.json
// capturing immediate and deferred release() events for traceability during debugging.
func writeIRRAIIDebug(pkg, unit string, f *ast.File) (string, error) {
    if f == nil { return "", nil }
    type pos struct{ Line, Column, Offset int }
    type event struct{ Kind, Target string; Pos pos }
    type fn struct{ Name string; Releases []event }
    obj := map[string]any{"schema": "ir.raii.v1", "package": pkg, "unit": unit, "functions": []any{}}
    var fns []any
    for _, d := range f.Decls {
        fd, ok := d.(*ast.FuncDecl)
        if !ok || fd.Body == nil { continue }
        var evs []event
        // Walk statements to collect immediate and deferred release() targets.
        for _, st := range fd.Body.Stmts {
            switch v := st.(type) {
            case *ast.DeferStmt:
                if name, p := raiiTargetFromCall(v.Call); name != "" {
                    evs = append(evs, event{Kind: "defer", Target: name, Pos: pos{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                }
            case *ast.ExprStmt:
                if name, p := raiiTargetFromExpr(v.X); name != "" {
                    evs = append(evs, event{Kind: "immediate", Target: name, Pos: pos{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                }
            }
        }
        fns = append(fns, fn{Name: fd.Name, Releases: evs})
    }
    obj["functions"] = fns
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, _ := json.MarshalIndent(obj, "", "  ")
    out := filepath.Join(dir, unit+".raii.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
// helpers moved to separate files to satisfy single-declaration rule
