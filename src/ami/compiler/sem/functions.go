package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeFunctions checks basic function semantics: no duplicates; no blank names or parameters.
func AnalyzeFunctions(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    seen := map[string]struct{}{}
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok { continue }
        // function name cannot be blank identifier
        if fn.Name == "_" || fn.Name == "" {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_BLANK_IDENT_ILLEGAL", Message: "blank identifier not allowed for function name", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
        }
        if _, exists := seen[fn.Name]; exists {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DUP_FUNC", Message: "duplicate function name", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
        } else {
            if fn.Name != "" { seen[fn.Name] = struct{}{} }
        }
        // parameter names cannot be blank identifier
        for _, p := range fn.Params {
            if p.Name == "_" || p.Name == "" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_BLANK_PARAM_ILLEGAL", Message: "blank identifier not allowed for parameter name", Pos: &diag.Position{Line: p.Pos.Line, Column: p.Pos.Column, Offset: p.Pos.Offset}})
            }
        }
    }
    return out
}

