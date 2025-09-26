package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// analyzeMemoryDomains enforces basic allocation domain separation (6.5/Ch.2.4):
// - Event heap (Event<T>), Node-state (State), Ephemeral stack (locals/others).
// Forbidden cross-domain references:
//   - Assigning address of non-state value into state memory, e.g., `*st = &ev` or `*st = &x`.
//     Emits E_CROSS_DOMAIN_REF.
func analyzeMemoryDomains(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Body) == 0 {
        return diags
    }
    // Token-based scan for prohibited cross-domain patterns. We prefer token
    // matching to avoid coupling to expression forms and to work even when the
    // parser recorded address-of errors.
    env := map[string]astpkg.TypeRef{}
    for _, p := range fd.Params {
        if p.Name != "" {
            env[p.Name] = p.Type
        }
    }
    toks := fd.Body
    for i := 0; i+3 < len(toks); i++ {
        // *st = &x (address-of into state)
        if toks[i].Kind == tok.STAR && toks[i+1].Kind == tok.IDENT && toks[i+2].Kind == tok.ASSIGN && toks[i+3].Kind == tok.AMP {
            if tr, ok := env[toks[i+1].Lexeme]; ok && tr.Name == "State" {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CROSS_DOMAIN_REF", Message: "cross-domain reference into state: cannot assign address-of non-state value into state"})
            }
        }
        // *st = ev (assignment from non-state identifier into state)
        if toks[i].Kind == tok.STAR && toks[i+1].Kind == tok.IDENT && toks[i+2].Kind == tok.ASSIGN && toks[i+3].Kind == tok.IDENT {
            if tr, ok := env[toks[i+1].Lexeme]; ok && tr.Name == "State" {
                if rt, ok2 := env[toks[i+3].Lexeme]; ok2 && rt.Name != "State" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CROSS_DOMAIN_REF", Message: "cross-domain assignment into state from non-state value is forbidden"})
                }
            }
        }
    }
    return diags
}

