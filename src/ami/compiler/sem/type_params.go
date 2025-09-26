package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
)

// analyzeFuncTypeParams enforces minimal generic parameter rules:
// - Reject duplicate type parameter names within the same function.
func analyzeFuncTypeParams(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    seen := map[string]bool{}
    for _, tp := range fd.TypeParams {
        if tp.Name == "" { continue }
        if seen[tp.Name] {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_TYPE_PARAM", Message: "duplicate type parameter name in function"})
            // continue scanning to surface all duplicates in one pass
            continue
        }
        seen[tp.Name] = true
    }
    return diags
}

