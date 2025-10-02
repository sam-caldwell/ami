package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeWorkers validates worker references in pipeline Transform nodes.
// - Emits E_WORKER_UNDEFINED when referenced worker function is not declared in the file (simple, local scope only).
// - Emits E_WORKER_SIGNATURE when a non-factory worker does not match the canonical signature
//   func(Event<T>) (Event<U>, error) under text-based checks.
func AnalyzeWorkers(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // index local functions by name
    funcs := map[string]*ast.FuncDecl{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            funcs[fn.Name] = fn
        }
    }
    // collect import aliases for dotted worker suppression
    imports := ImportAliases(f)
    // scan pipelines and their steps
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            if strings.ToLower(st.Name) != "transform" { continue }
            if len(st.Args) == 0 { continue }
            a0 := st.Args[0]
            wname := a0.Text
            if a0.IsString {
                // use literal value
                wname = a0.Text
            }
            // trim quotes if present
            wname = strings.Trim(wname, "\"")
            if wname == "" { continue }
            if i := strings.LastIndexByte(wname, '.'); i >= 0 {
                // if prefix is an import alias, treat as external and skip undefined/signature checks
                prefix := wname[:i]
                if _, ok := imports[prefix]; ok {
                    continue
                }
                wname = wname[i+1:]
            }
            fn, ok := funcs[wname]
            if !ok {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_UNDEFINED", Message: "worker undefined: " + wname, Pos: &diag.Position{Line: a0.Pos.Line, Column: a0.Pos.Column, Offset: a0.Pos.Offset}})
                continue
            }
            // Workers cannot be decorated
            if len(fn.Decorators) > 0 {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_ON_WORKER", Message: "workers cannot be decorated", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
            }
            // accept factories New* without signature checks
            if strings.HasPrefix(fn.Name, "New") { continue }
            // signature check: 1 param Event<...>; 2 results: Event<...>, error
            if len(fn.Params) != 1 || !strings.HasPrefix(fn.Params[0].Type, "Event<") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: "invalid worker signature: want func(Event<T>)->(Event<U>, error)", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                if len(fn.Decorators) > 0 {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_SIGNATURE", Message: "decorators must not change worker signature", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                }
                continue
            }
            if len(fn.Results) != 2 || !strings.HasPrefix(fn.Results[0].Type, "Event<") || fn.Results[1].Type != "error" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: "invalid worker signature: want func(Event<T>)->(Event<U>, error)", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                if len(fn.Decorators) > 0 {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_SIGNATURE", Message: "decorators must not change worker signature", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                }
                continue
            }
            // Enforce pointer-free Event<T>/Event<U> type arguments.
            if t := eventTypeArg(fn.Params[0].Type); t != "" {
                if strings.ContainsAny(t, "&*") {
                    p := fn.Params[0].Pos
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_PTR_FORBIDDEN", Message: "Event<T> must be pointer-free (no '&' or '*')", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                }
            }
            if t := eventTypeArg(fn.Results[0].Type); t != "" {
                if strings.ContainsAny(t, "&*") {
                    p := fn.Results[0].Pos
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_PTR_FORBIDDEN", Message: "Event<U> must be pointer-free (no '&' or '*')", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                }
            }
        }
    }
    return out
}

// eventTypeArg extracts the inner type parameter from an Event<...> type string.
// Returns empty string if not in the expected form.
