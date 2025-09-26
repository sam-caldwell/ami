package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

func analyzeCallTypes(fd astpkg.FuncDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.BodyStmts) == 0 {
        return diags
    }
    // environment: parameter identifier -> TypeRef
    env := map[string]astpkg.TypeRef{}
    for _, p := range fd.Params {
        if p.Name != "" {
            env[p.Name] = p.Type
        }
    }

    // infer expression type where possible
    var exprType func(astpkg.Expr) (astpkg.TypeRef, bool)
    exprType = func(e astpkg.Expr) (astpkg.TypeRef, bool) {
        switch v := e.(type) {
        case astpkg.Ident:
            if tr, ok := env[v.Name]; ok {
                return tr, true
            }
            return astpkg.TypeRef{}, false
        case astpkg.ContainerLit:
            switch v.Kind {
            case "slice":
                var elem astpkg.TypeRef
                if len(v.TypeArgs) == 1 {
                    elem = v.TypeArgs[0]
                } else if len(v.Elems) > 0 {
                    if t, ok := exprType(v.Elems[0]); ok {
                        elem = t
                    }
                }
                return astpkg.TypeRef{Name: "slice", Args: []astpkg.TypeRef{elem}}, true
            case "set":
                var elem astpkg.TypeRef
                if len(v.TypeArgs) == 1 {
                    elem = v.TypeArgs[0]
                } else if len(v.Elems) > 0 {
                    if t, ok := exprType(v.Elems[0]); ok {
                        elem = t
                    }
                }
                return astpkg.TypeRef{Name: "set", Args: []astpkg.TypeRef{elem}}, true
            case "map":
                var kt, vt astpkg.TypeRef
                if len(v.TypeArgs) == 2 {
                    kt, vt = v.TypeArgs[0], v.TypeArgs[1]
                } else if len(v.MapElems) > 0 {
                    if t1, ok := exprType(v.MapElems[0].Key); ok {
                        kt = t1
                    }
                    if t2, ok := exprType(v.MapElems[0].Value); ok {
                        vt = t2
                    }
                }
                return astpkg.TypeRef{Name: "map", Args: []astpkg.TypeRef{kt, vt}}, true
            }
            return astpkg.TypeRef{}, false
        case astpkg.BasicLit:
            if v.Kind == "string" {
                return astpkg.TypeRef{Name: "string"}, true
            }
            if v.Kind == "number" {
                return astpkg.TypeRef{Name: "int"}, true
            }
            return astpkg.TypeRef{}, false
        case astpkg.UnaryExpr:
            if v.Op == "*" {
                return exprType(v.X)
            }
            return astpkg.TypeRef{}, false
        case astpkg.BinaryExpr:
            // derive simple operator result types
            lt, lok := exprType(v.X)
            rt, rok := exprType(v.Y)
            if !(lok && rok) {
                return astpkg.TypeRef{}, false
            }
            switch v.Op {
            case "+", "-", "*", "/", "%":
                if strings.ToLower(lt.Name) == "int" && strings.ToLower(rt.Name) == "int" {
                    return astpkg.TypeRef{Name: "int"}, true
                }
                return astpkg.TypeRef{}, false
            case "==", "!=", "<", "<=", ">", ">=":
                return astpkg.TypeRef{Name: "bool"}, true
            }
            return astpkg.TypeRef{}, false
        case astpkg.CallExpr:
            // If callee is an identifier and we know its declaration, and it returns exactly one type, propagate that.
            switch c := v.Fun.(type) {
            case astpkg.Ident:
                if decl, ok := funcs[c.Name]; ok {
                    if len(decl.Result) == 1 {
                        return decl.Result[0], true
                    }
                }
            case astpkg.SelectorExpr:
                // Unknown for now
            }
            return astpkg.TypeRef{}, false
        default:
            return astpkg.TypeRef{}, false
        }
    }

    // Simple structural unify with single-letter type variables
    type substMap = map[string]astpkg.TypeRef
    var isTypeVar = func(name string) bool {
        if len(name) != 1 {
            return false
        }
        b := name[0]
        return b >= 'A' && b <= 'Z'
    }
    var unify func(want, got astpkg.TypeRef, subst substMap) bool
    unify = func(want, got astpkg.TypeRef, subst substMap) bool {
        // pointer and slice flags must match exactly
        if want.Ptr != got.Ptr || want.Slice != got.Slice {
            return false
        }
        // Generic variable binding
        if isTypeVar(want.Name) && len(want.Args) == 0 {
            if bound, ok := subst[want.Name]; ok {
                return typeRefToString(bound) == typeRefToString(got)
            }
            subst[want.Name] = got
            return true
        }
        // Names must match case-insensitively and arity must match
        if strings.ToLower(want.Name) != strings.ToLower(got.Name) {
            return false
        }
        if len(want.Args) != len(got.Args) {
            return false
        }
        for i := range want.Args {
            if !unify(want.Args[i], got.Args[i], subst) {
                return false
            }
        }
        return true
    }

    // Walk statements (track local var decls) and check calls
    var walkStmt func(astpkg.Stmt)
    var walkExpr func(astpkg.Expr)
    walkExpr = func(e astpkg.Expr) {
        switch v := e.(type) {
        case astpkg.CallExpr:
            // Only check calls to known local functions
            if id, ok := v.Fun.(astpkg.Ident); ok {
                if decl, ok2 := funcs[id.Name]; ok2 {
                    // Arity check
                    if len(v.Args) != len(decl.Params) {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CALL_ARITY_MISMATCH", Message: "function call arity mismatch"})
                        // still try to compare what we can
                    }
                    n := len(v.Args)
                    if len(decl.Params) < n {
                        n = len(decl.Params)
                    }
                    // reset substitution per call
                    subst := make(substMap)
                    for i := 0; i < n; i++ {
                        at, aok := exprType(v.Args[i])
                        if !aok {
                            // if parameter expects a type variable (directly or as generic arg), mark ambiguous
                            pt := decl.Params[i].Type
                            if hasTypeVar(pt) {
                                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_AMBIGUOUS", Message: "cannot infer generic type argument from call site"})
                            }
                            continue
                        }
                        pt := decl.Params[i].Type
                        if !unify(pt, at, subst) {
                            msg := "call argument type mismatch: got " + typeRefToString(at) + ", want " + typeRefToString(pt)
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CALL_ARG_TYPE_MISMATCH", Message: msg})
                        }
                    }
                }
            } else {
                // nested selector/unknown callee: still walk args
            }
            for _, a := range v.Args {
                walkExpr(a)
            }
        case astpkg.UnaryExpr:
            walkExpr(v.X)
        case astpkg.SelectorExpr:
            walkExpr(v.X)
        case astpkg.Expr:
            // nothing else to walk
            _ = v
        }
    }
    walkStmt = func(s astpkg.Stmt) {
        switch v := s.(type) {
        case astpkg.VarDeclStmt:
            if v.Type.Name != "" {
                env[v.Name] = v.Type
            } else if v.Init != nil {
                if t, ok := exprType(v.Init); ok {
                    env[v.Name] = t
                }
            }
        case astpkg.AssignStmt:
            walkExpr(v.LHS)
            walkExpr(v.RHS)
        case astpkg.ExprStmt:
            walkExpr(v.X)
        case astpkg.BlockStmt:
            for _, ss := range v.Stmts {
                walkStmt(ss)
            }
        case astpkg.DeferStmt:
            if v.X != nil {
                walkExpr(v.X)
            }
        case astpkg.MutBlockStmt:
            for _, ss := range v.Body.Stmts {
                walkStmt(ss)
            }
        case astpkg.ReturnStmt:
            for _, r := range v.Results {
                walkExpr(r)
            }
        default:
            // nothing
        }
    }
    for _, s := range fd.BodyStmts {
        walkStmt(s)
    }
    return diags
}

func hasTypeVar(t astpkg.TypeRef) bool {
    if len(t.Name) == 1 && t.Name[0] >= 'A' && t.Name[0] <= 'Z' {
        return true
    }
    for _, a := range t.Args {
        if hasTypeVar(a) {
            return true
        }
    }
    return false
}

