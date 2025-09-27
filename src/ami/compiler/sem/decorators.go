package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeDecorators performs basic resolution and consistency checks for function decorators.
// Scaffold rules:
// - Built-ins: deprecated, metrics (recognized without requiring a top-level symbol)
// - Otherwise, decorator name must resolve to a top-level function declared in the same file
// - Duplicate decorator with different arg list emits E_DECORATOR_CONFLICT
// - Unresolved decorator emits E_DECORATOR_UNRESOLVED
func AnalyzeDecorators(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // index top-level function names
    funcs := map[string]struct{}{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            funcs[fn.Name] = struct{}{}
        }
    }
    // known built-ins
    builtins := map[string]struct{}{"deprecated": {}, "metrics": {}}
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok { continue }
        if len(fn.Decorators) == 0 { continue }
        seen := map[string]string{}
        for _, dec := range fn.Decorators {
            name := strings.TrimSpace(dec.Name)
            if name == "" { // unresolved empty
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_UNRESOLVED", Message: "decorator name is empty", Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name}})
                continue
            }
            // allow dotted names but resolve by last segment for local function resolution
            base := name
            if i := strings.LastIndexByte(name, '.'); i >= 0 && i+1 < len(name) {
                base = name[i+1:]
            }
            // resolution: built-in or local function exists
            if _, ok := builtins[name]; !ok { // exact built-in match
                if _, ok := builtins[base]; !ok { // allow short built-in names too
                    if _, ok := funcs[base]; !ok {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DECORATOR_UNRESOLVED", Message: "decorator unresolved: " + name, Pos: &diag.Position{Line: dec.Pos.Line, Column: dec.Pos.Column, Offset: dec.Pos.Offset}, Data: map[string]any{"function": fn.Name}})
                    }
                }
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

func canonicalDecoArgs(args []ast.Arg) string {
    if len(args) == 0 { return "" }
    // Preserve order; join with '|'
    out := make([]string, 0, len(args))
    for _, a := range args { out = append(out, a.Text) }
    return strings.Join(out, "|")
}

