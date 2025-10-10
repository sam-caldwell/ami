package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeWorkers validates worker references in pipeline nodes.
// It supports both positional and named-argument styles:
//   - Transform FooWorker
//   - Transform(worker=FooWorker)
//   - Ingress(worker=...)
//   - Egress(worker=...)
// It will:
// - Emit E_WORKER_UNDEFINED when a referenced local worker function is not declared in the file.
// - Emit E_WORKER_SIGNATURE when a non-factory worker does not match the canonical signature
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
            // Only consider steps that can carry a worker
            switch strings.ToLower(st.Name) {
            case "transform", "ingress", "egress":
            default:
                continue
            }

            // Resolve worker name/value from args:
            // - prefer named arg 'worker=...'
            // - else if first positional arg is present and not a key=value, treat as worker
            // - if value looks like a function literal 'func(' then accept without lookup
            var warg ast.Arg
            var hasWorker bool
            for _, a := range st.Args {
                if eq := strings.IndexByte(a.Text, '='); eq > 0 {
                    key := strings.TrimSpace(a.Text[:eq])
                    if strings.EqualFold(key, "worker") {
                        // capture only the right-hand side
                        rhs := strings.TrimSpace(a.Text[eq+1:])
                        warg = ast.Arg{Pos: a.Pos, Text: rhs, IsString: a.IsString}
                        hasWorker = true
                        break
                    }
                    continue
                }
                // positional argument encountered; treat first positional as candidate worker
                if !hasWorker {
                    warg = a
                    hasWorker = true
                }
            }
            if !hasWorker { continue }

            wname := strings.Trim(warg.Text, "\"")
            if wname == "" { continue }
            // Inline function literal? Parse and validate signature and accept.
            if wname == "func" || strings.HasPrefix(wname, "func") {
                // Try to parse a signature from the literal text
                paramTyp, results, ok := inlineFuncSig(strings.TrimSpace(warg.Text))
                if !ok {
                    // Unable to parse; treat as soft acceptance to avoid false positives.
                    continue
                }
                // Enforce pointer-free Event<...> for param/result when present
                if t := eventTypeArg(paramTyp); t != "" {
                    if strings.ContainsAny(t, "&*") {
                        p := warg.Pos
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_PTR_FORBIDDEN", Message: "Event<T> must be pointer-free (no '&' or '*')", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                    }
                }
                if len(results) > 0 {
                    if t := eventTypeArg(results[0]); t != "" {
                        if strings.ContainsAny(t, "&*") {
                            p := warg.Pos
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_PTR_FORBIDDEN", Message: "Event<U> must be pointer-free (no '&' or '*')", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
                        }
                    }
                }
                // Basic shape acceptance: either single 'error' or two results with second 'error'.
                if len(results) == 1 {
                    if strings.TrimSpace(results[0]) != "error" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: "invalid worker signature: want error or (X, error)", Pos: &diag.Position{Line: warg.Pos.Line, Column: warg.Pos.Column, Offset: warg.Pos.Offset}})
                    }
                } else if len(results) >= 2 {
                    if strings.TrimSpace(results[1]) != "error" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: "invalid worker signature: second result must be error", Pos: &diag.Position{Line: warg.Pos.Line, Column: warg.Pos.Column, Offset: warg.Pos.Offset}})
                    }
                } else {
                    // No results at all: treat as invalid
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: "invalid worker signature: missing results", Pos: &diag.Position{Line: warg.Pos.Line, Column: warg.Pos.Column, Offset: warg.Pos.Offset}})
                }
                continue
            }
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
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_UNDEFINED", Message: "worker undefined: " + wname, Pos: &diag.Position{Line: warg.Pos.Line, Column: warg.Pos.Column, Offset: warg.Pos.Offset}})
                continue
            }
            // Workers cannot be decorated
            if len(fn.Decorators) > 0 {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_ON_WORKER", Message: "workers cannot be decorated", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
            }
            // accept factories New* without signature checks
            if strings.HasPrefix(fn.Name, "New") { continue }
            // signature check (docx-aligned): 1 param Event<...>; 2 results with second 'error'; first result may be U or Event<U>.
            if len(fn.Params) != 1 || !strings.HasPrefix(fn.Params[0].Type, "Event<") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: "invalid worker signature: want func(Event<T>)->(U, error) or (Event<U>, error)", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                if len(fn.Decorators) > 0 {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_SIGNATURE", Message: "decorators must not change worker signature", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                }
                continue
            }
            // Two results with second 'error'
            if len(fn.Results) != 2 || fn.Results[1].Type != "error" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: "invalid worker signature: want func(Event<T>)->(U, error) or (Event<U>, error)", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                if len(fn.Decorators) > 0 {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_SIGNATURE", Message: "decorators must not change worker signature", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
                }
                continue
            }
            // Permit first result as U or Event<U>; enforce pointer-free when Event<...>.
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
