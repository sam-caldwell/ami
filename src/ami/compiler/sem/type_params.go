package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    srcset "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// analyzeFuncTypeParams enforces minimal generic parameter rules:
// - Reject duplicate type parameter names within the same function.
func analyzeFuncTypeParams(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    seen := map[string]bool{}
    for _, tp := range fd.TypeParams {
        if tp.Name == "" { continue }
        if seen[tp.Name] {
            // Attach function start position (best available without per-param offsets)
            pos := &srcset.Position{Line: fd.Pos.Line, Column: fd.Pos.Column, Offset: fd.Pos.Offset}
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_TYPE_PARAM", Message: "duplicate type parameter name in function", Pos: pos})
            // continue scanning to surface all duplicates in one pass
            continue
        }
        seen[tp.Name] = true
    }
    return diags
}
