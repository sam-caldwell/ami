package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// analyzeMutationMarkers enforces AMI mutability rules:
// - No Rust-like `mut { ... }` blocks are permitted.
// - Any assignment must use `*` on the left-hand side to mark mutation.
// - Unary '*' is not a dereference and is invalid in expression (RHS) position.
// Note: In AMI 2.3.2, `*` on the LHS is the mutation marker (not a pointer
// dereference). This function validates that usage and flags misuse.
func analyzeMutationMarkers(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Body) == 0 && len(fd.BodyStmts) == 0 {
        return diags
    }
    if len(fd.BodyStmts) > 0 {
        var walkExpr func(astpkg.Expr, bool)
        walkExpr = func(e astpkg.Expr, isLHS bool) {
            switch v := e.(type) {
            case astpkg.UnaryExpr:
                if v.Op == "*" {
                    if !isLHS {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STAR_MISUSED", Message: "'*' is not a dereference; only allowed on assignment left-hand side as a mutability marker"})
                    }
                    walkExpr(v.X, isLHS)
                }
                // '&' is handled by parser; ignore here
            case astpkg.CallExpr:
                for _, a := range v.Args {
                    walkExpr(a, false)
                }
            case astpkg.SelectorExpr:
                walkExpr(v.X, false)
            }
        }
        var walkStmt func(astpkg.Stmt)
        walkStmt = func(s astpkg.Stmt) {
            switch v := s.(type) {
            case astpkg.AssignStmt:
                if ue, ok := v.LHS.(astpkg.UnaryExpr); !ok || ue.Op != "*" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_ASSIGN_UNMARKED", Message: "assignment must be marked with '*' (mutation marker) on left-hand side"})
                }
                walkExpr(v.RHS, false)
            case astpkg.MutBlockStmt:
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_BLOCK_UNSUPPORTED", Message: "mut { ... } blocks are not part of AMI; use mutate(expr) or '*' on assignment LHS"})
                for _, ss := range v.Body.Stmts {
                    walkStmt(ss)
                }
            case astpkg.BlockStmt:
                for _, ss := range v.Stmts {
                    walkStmt(ss)
                }
            case astpkg.ExprStmt:
                walkExpr(v.X, false)
            }
        }
        for _, s := range fd.BodyStmts {
            walkStmt(s)
        }
        return diags
    }
    // Token-based fallback
    for _, t := range fd.Body {
        if t.Kind == tok.KW_MUT {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_BLOCK_UNSUPPORTED", Message: "mut { ... } blocks are not part of AMI; use mutate(expr) or '*' on assignment LHS"})
        }
    }
    toks := fd.Body
    for i := 0; i < len(toks); i++ {
        if toks[i].Kind == tok.ASSIGN {
            if !(i-2 >= 0 && toks[i-2].Kind == tok.STAR && toks[i-1].Kind == tok.IDENT) {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_ASSIGN_UNMARKED", Message: "assignment must be marked with '*' (mutation marker) on left-hand side"})
            }
        }
    }
    return diags
}

