package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// analyzeEventContracts enforces event parameter immutability: the Event<T>
// parameter (commonly named 'ev') cannot be assigned.
func analyzeEventContracts(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Params) < 1 || len(fd.Body) == 0 {
        return diags
    }
    // detect event parameter name
    evName := ""
    // canonical: first parameter is Event<T>
    if p := fd.Params[0]; p.Type.Name == "Event" && len(p.Type.Args) == 1 && p.Name != "" {
        evName = p.Name
    }
    if evName == "" {
        return diags
    }
    toks := fd.Body
    for i := 0; i < len(toks); i++ {
        if toks[i].Kind == tok.ASSIGN {
            // LHS ident equal to event param, not a deref
            if i-1 >= 0 && toks[i-1].Kind == tok.IDENT && toks[i-1].Lexeme == evName {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EVENT_PARAM_ASSIGN", Message: "event parameter is immutable and cannot be reassigned"})
            }
        }
    }
    return diags
}
