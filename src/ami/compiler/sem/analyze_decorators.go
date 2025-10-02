package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

func AnalyzeDecorators(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // recognize canonical worker signature
    typeIsEvent := func(s string) bool { return s == "Event" || strings.HasPrefix(s, "Event<") }
    isWorkerSig := func(fn *ast.FuncDecl) bool {
        if fn == nil { return false }
        if len(fn.Params) != 1 { return false }
        if !typeIsEvent(fn.Params[0].Type) { return false }
        if len(fn.Results) != 2 { return false }
        if !typeIsEvent(fn.Results[0].Type) { return false }
        if fn.Results[1].Type != "error" { return false }
        return true
    }
    // index top-level function names
    funcs := map[string]struct{}{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok { funcs[fn.Name] = struct{}{} }
    }
    // known built-ins
    builtins := map[string]struct{}{"deprecated": {}, "metrics": {}}
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok { continue }
        if len(fn.Decorators) == 0 { continue }
        // reject decorators on worker functions
        if isWorkerSig(fn) {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_ON_WORKER", Message: "workers cannot be decorated", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}, Data: map[string]any{"function": fn.Name}})
        }
        seen := map[string]string{}
        for _, dec := range fn.Decorators {
            name := strings.TrimSpace(dec.Name)
            if name == "" { // unresolved empty
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_UNRESOLVED", Message: "decorator name is empty", Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name}})
                continue
            }
            // allow dotted names but resolve by last segment for local function resolution
            base := name
            if i := strings.LastIndexByte(name, '.'); i >= 0 && i+1 < len(name) { base = name[i+1:] }
            // disabled check
            if _, off := disabledDecorators[name]; off { out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_DISABLED", Message: "decorator disabled: " + name, Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name}}) }
            if _, off := disabledDecorators[base]; off { out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_DISABLED", Message: "decorator disabled: " + base, Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name}}) }
            // resolution: built-in or local function exists
            if _, ok := builtins[name]; !ok { // exact built-in match
                if _, ok := builtins[base]; !ok { // allow short built-in names too
                    if _, ok := funcs[base]; !ok {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_UNDEFINED", Message: "decorator undefined: " + name, Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name}})
                    }
                }
            }
            // Built-in behaviors (scaffold)
            if name == "deprecated" || base == "deprecated" {
                msg := canonicalDecoArgs(dec.Args)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_DEPRECATED", Message: "function is deprecated", Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name, "message": msg}})
            }
            // detect conflicting duplicates
            val := canonicalDecoArgs(dec.Args)
            if prev, ok := seen[name]; ok && prev != val {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_CONFLICT", Message: "conflicting decorator arguments: " + name, Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name}})
            }
            seen[name] = val
        }
    }
    return out
}

