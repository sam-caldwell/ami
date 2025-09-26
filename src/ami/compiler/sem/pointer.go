package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// analyzePointerProhibitions enforces AMI 2.3.2: no raw pointer/address operators.
// Specifically, disallow '&' anywhere in function bodies.
func analyzePointerProhibitions(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Body) == 0 {
        return diags
    }
    for _, t := range fd.Body {
        if t.Kind == tok.AMP {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "'&' address-of operator is not allowed; AMI does not expose raw pointers (see 2.3.2)"})
        }
    }
    return diags
}

