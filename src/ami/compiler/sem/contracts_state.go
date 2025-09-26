package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// analyzeStateContracts enforces basic node-state parameter rules:
// - State parameter is immutable (cannot be reassigned): E_STATE_PARAM_ASSIGN
// - Ambient state: pointer parameters to State are forbidden: E_STATE_PARAM_POINTER
//   (use ambient state.get/set/update/list per SPEC §6.3, AMI 2.3.2)
// - No address-of in AMI 2.3.2
func analyzeStateContracts(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Params) < 1 {
        return diags
    }
    // Explicitly reject pointer State parameters (even if parser already flagged generic pointer usage)
    for _, p := range fd.Params {
        if p.Type.Name == "State" && p.Type.Ptr {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STATE_PARAM_POINTER", Message: "state is ambient; do not take *State parameters. Use state.get/set/update/list"})
        }
    }
    if len(fd.Body) == 0 {
        return diags
    }
    // Build simple env of param name -> type
    env := map[string]astpkg.TypeRef{}
    for _, p := range fd.Params {
        if p.Name != "" {
            env[p.Name] = p.Type
        }
    }
    toks := fd.Body
    for i := 0; i < len(toks); i++ {
        // Reassignment of state param: IDENT '=' ... where IDENT is State
        if toks[i].Kind == tok.IDENT {
            if tr, ok := env[toks[i].Lexeme]; ok && tr.Name == "State" {
                if i+1 < len(toks) && toks[i+1].Kind == tok.ASSIGN {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STATE_PARAM_ASSIGN", Message: "state parameter is immutable and cannot be reassigned"})
                }
            }
        }
        // No address-of in AMI 2.3.2
    }
    return diags
}
