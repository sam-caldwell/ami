package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// analyzeOperators scans tokens to validate basic arithmetic and comparison operand types.
// Arithmetic: +,-,*,/,% expect numeric (int) on both sides.
// Comparison: ==,!= require same types; <,<=,>,>= allowed for int (and strings not yet guaranteed).
func analyzeOperators(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // If AST is available, operator checks are handled via exprType on BinaryExpr
    if len(fd.BodyStmts) > 0 {
        return diags
    }
    if len(fd.Body) == 0 {
        return diags
    }
    // build env from params
    env := map[string]astpkg.TypeRef{}
    for _, p := range fd.Params {
        if p.Name != "" {
            env[p.Name] = p.Type
        }
    }
    // helpers
    resolve := func(t tok.Token) (astpkg.TypeRef, bool) {
        switch t.Kind {
        case tok.IDENT:
            if tr, ok := env[t.Lexeme]; ok {
                return tr, true
            }
        case tok.STRING:
            return astpkg.TypeRef{Name: "string"}, true
        case tok.NUMBER:
            return astpkg.TypeRef{Name: "int"}, true
        }
        return astpkg.TypeRef{}, false
    }
    same := func(a, b astpkg.TypeRef) bool { return typeRefToString(a) == typeRefToString(b) }
    toks := fd.Body
    // simple binary op scanning: IDENT/STRING/NUMBER op IDENT/STRING/NUMBER
    // This is a token-level heuristic and not a full expression parser.
    for i := 1; i+1 < len(toks); i++ {
        op := toks[i]
        switch op.Kind {
        case tok.EQ, tok.NEQ, tok.LT, tok.LTE, tok.GT, tok.GTE, tok.PLUS, tok.MINUS, tok.STAR, tok.SLASH, tok.PERCENT:
            lt, okL := resolve(toks[i-1])
            rt, okR := resolve(toks[i+1])
            if !okL || !okR {
                continue
            }
            switch op.Kind {
            case tok.EQ, tok.NEQ:
                if !same(lt, rt) {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "comparison operands must have the same type"})
                }
            case tok.LT, tok.LTE, tok.GT, tok.GTE:
                if lt.Name != "int" || rt.Name != "int" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "ordering comparisons require integer operands"})
                }
            case tok.PLUS, tok.MINUS, tok.STAR, tok.SLASH, tok.PERCENT:
                if lt.Name != "int" || rt.Name != "int" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "arithmetic operands must be integers"})
                }
            }
        }
    }
    return diags
}
