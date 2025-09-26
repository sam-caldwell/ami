package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    srcset "github.com/sam-caldwell/ami/src/ami/compiler/source"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "strings"
)

// analyzeImperativeTypes performs minimal type checking over function bodies by
// scanning tokens and leveraging parameter types as an environment.
// Supported checks:
//   - E_DEREF_TYPE: '*' applied to non-pointer parameter identifier.
//   - E_ASSIGN_TYPE_MISMATCH: for simple forms `x = y`, `*p = y`, `x = &y` when
//     both sides resolve to known types from parameters or literals.
func analyzeImperativeTypes(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Body) == 0 {
        return diags
    }
    if len(fd.BodyStmts) > 0 {
        // AST-based analysis is handled by AnalyzeFile via analyzeImperativeTypesFromAST with funcs context
        return diags
    }
    // env maps parameter identifiers to their TypeRef
    env := map[string]astpkg.TypeRef{}
    for _, p := range fd.Params {
        if p.Name != "" {
            env[p.Name] = p.Type
        }
    }
    // helpers
    typeStr := func(tr astpkg.TypeRef) string { return typeRefToString(tr) }
    // resolve expression type starting at token index i; returns type string and ok
    // Supports IDENT, '*' IDENT, '&' IDENT, STRING→string, NUMBER→int
    resolve := func(toks []tok.Token, i int) (string, bool, bool) {
        // third return is "hardError" indicator already emitted; used to avoid double-diags
        if i >= len(toks) {
            return "", false, false
        }
        switch toks[i].Kind {
        case tok.IDENT:
            if tr, ok := env[toks[i].Lexeme]; ok {
                return typeStr(tr), true, false
            }
            return "", false, false
        case tok.STAR, tok.AMP:
            // no raw pointer semantics in AMI
            return "", false, false
        case tok.STRING:
            return "string", true, false
        case tok.NUMBER:
            return "int", true, false
        default:
            return "", false, false
        }
    }
    toks := fd.Body
    for i := 0; i < len(toks); i++ {
        // Simple assignment patterns: [IDENT|* IDENT] '=' <expr>
        if toks[i].Kind == tok.ASSIGN {
            // LHS type
            var lhs string
            var okL bool
            if !okL && i-1 >= 0 && toks[i-1].Kind == tok.IDENT {
                if tr, ok := env[toks[i-1].Lexeme]; ok {
                    lhs, okL = typeStr(tr), true
                }
            }
            // RHS type
            rhs, okR, _ := resolve(toks, i+1)
            if okL && okR && lhs != rhs {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ASSIGN_TYPE_MISMATCH", Message: "assignment type mismatch: " + lhs + " != " + rhs})
            }
        }
    }
    return diags
}

func analyzeImperativeTypesFromAST(fd astpkg.FuncDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    env := map[string]astpkg.TypeRef{}
    for _, p := range fd.Params {
        if p.Name != "" {
            env[p.Name] = p.Type
        }
    }
    typeStr := func(tr astpkg.TypeRef) string { return typeRefToString(tr) }
    // single-letter type variable helper & unifier
    isTypeVar := func(name string) bool { return len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' }
    var unify func(want, got astpkg.TypeRef, subst map[string]astpkg.TypeRef) bool
    unify = func(want, got astpkg.TypeRef, subst map[string]astpkg.TypeRef) bool {
        if want.Ptr != got.Ptr || want.Slice != got.Slice {
            return false
        }
        if isTypeVar(want.Name) && len(want.Args) == 0 {
            if b, ok := subst[want.Name]; ok {
                return typeStr(b) == typeStr(got)
            }
            subst[want.Name] = got
            return true
        }
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
    // applySubst replaces single-letter type variables in a TypeRef using subst map
    var applySubst func(astpkg.TypeRef, map[string]astpkg.TypeRef) astpkg.TypeRef
    applySubst = func(t astpkg.TypeRef, subst map[string]astpkg.TypeRef) astpkg.TypeRef {
        if len(t.Name) == 1 && t.Name[0] >= 'A' && t.Name[0] <= 'Z' && len(t.Args) == 0 {
            if b, ok := subst[t.Name]; ok {
                return b
            }
        }
        if len(t.Args) > 0 {
            out := t
            out.Args = make([]astpkg.TypeRef, len(t.Args))
            for i, a := range t.Args {
                out.Args[i] = applySubst(a, subst)
            }
            return out
        }
        return t
    }

    var exprType func(astpkg.Expr, bool) (astpkg.TypeRef, bool)
    exprType = func(e astpkg.Expr, isLHS bool) (astpkg.TypeRef, bool) {
        switch v := e.(type) {
        case astpkg.Ident:
            if tr, ok := env[v.Name]; ok {
                return tr, true
            }
            return astpkg.TypeRef{}, false
        case astpkg.ContainerLit:
            switch v.Kind {
            case "slice", "set":
                // element type is TypeArgs[0] if provided; otherwise infer from elements
                var elemT astpkg.TypeRef
                hasType := len(v.TypeArgs) == 1
                if hasType {
                    elemT = v.TypeArgs[0]
                }
                subst := map[string]astpkg.TypeRef{}
                for _, el := range v.Elems {
                    et, ok := exprType(el, false)
                    if !ok {
                        continue
                    }
                    if hasType {
                        if !unify(elemT, et, subst) {
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "container element type mismatch"})
                        }
                    } else {
                        // first element determines type; subsequent must match
                        if elemT.Name == "" {
                            elemT = et
                        } else {
                            if !unify(elemT, et, subst) {
                                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "container element type mismatch"})
                            }
                        }
                    }
                }
                if v.Kind == "slice" {
                    return astpkg.TypeRef{Name: "slice", Args: []astpkg.TypeRef{elemT}}, true
                }
                return astpkg.TypeRef{Name: "set", Args: []astpkg.TypeRef{elemT}}, true
            case "map":
                // determine key/value types either from annotation or first elements,
                // and validate consistency of all map entries
                var kt, vt astpkg.TypeRef
                hasTypes := len(v.TypeArgs) == 2
                if hasTypes {
                    kt, vt = v.TypeArgs[0], v.TypeArgs[1]
                }
                if len(v.MapElems) > 0 {
                    if !hasTypes {
                        if t1, ok := exprType(v.MapElems[0].Key, false); ok {
                            kt = t1
                        }
                        if t2, ok := exprType(v.MapElems[0].Value, false); ok {
                            vt = t2
                        }
                    }
                    subst := map[string]astpkg.TypeRef{}
                    for _, me := range v.MapElems {
                        if t1, ok := exprType(me.Key, false); ok {
                            if !unify(kt, t1, subst) {
                                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map key type mismatch"})
                            }
                        }
                        if t2, ok := exprType(me.Value, false); ok {
                            if !unify(vt, t2, subst) {
                                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map value type mismatch"})
                            }
                        }
                    }
                }
                return astpkg.TypeRef{Name: "map", Args: []astpkg.TypeRef{kt, vt}}, true
            }
            return astpkg.TypeRef{}, false
        case astpkg.UnaryExpr:
            if v.Op == "*" {
                // '*' is a mutation marker; not a pointer dereference. Permit only on LHS.
                if !isLHS {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STAR_MISUSED", Message: "'*' is not a dereference; only allowed on assignment left-hand side as a mutability marker"})
                }
                return exprType(v.X, isLHS)
            }
            return astpkg.TypeRef{}, false
        case astpkg.BinaryExpr:
            // validate homogenous operand types for arithmetic/comparison and infer result type
            lt, lok := exprType(v.X, false)
            rt, rok := exprType(v.Y, false)
            if lok && rok {
                switch v.Op {
                case "+", "-", "*", "/", "%":
                    if strings.EqualFold(lt.Name, "int") && strings.EqualFold(rt.Name, "int") {
                        return astpkg.TypeRef{Name: "int"}, true
                    }
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "arithmetic operands must be integers"})
                    return astpkg.TypeRef{}, true
                case "==", "!=", "<", "<=", ">", ">=":
                    if typeRefToString(lt) != typeRefToString(rt) && (v.Op == "==" || v.Op == "!=") {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "comparison operands must have the same type"})
                    }
                    return astpkg.TypeRef{Name: "bool"}, true
                }
            }
            return astpkg.TypeRef{}, true
        case astpkg.CallExpr:
            // infer instantiated return type for known local functions with single result
            if id, ok := v.Fun.(astpkg.Ident); ok {
                if decl, ok := funcs[id.Name]; ok {
                    // Build substitution from params <- args
                    subst := map[string]astpkg.TypeRef{}
                    n := len(v.Args)
                    if len(decl.Params) < n { n = len(decl.Params) }
                    for i := 0; i < n; i++ {
                        if at, aok := exprType(v.Args[i], false); aok {
                            _ = unify(decl.Params[i].Type, at, subst)
                        }
                    }
                    if len(decl.Result) == 1 {
                        rt := applySubst(decl.Result[0], subst)
                        return rt, true
                    }
                }
            }
            return astpkg.TypeRef{}, false
        case astpkg.BasicLit:
            switch v.Kind {
            case "string":
                return astpkg.TypeRef{Name: "string"}, true
            case "number":
                return astpkg.TypeRef{Name: "int"}, true
            }
            return astpkg.TypeRef{}, false
        default:
            return astpkg.TypeRef{}, false
        }
    }
    var walkStmt func(astpkg.Stmt)
    walkStmt = func(s astpkg.Stmt) {
        if as, ok := s.(astpkg.AssignStmt); ok {
            lt, lok := exprType(as.LHS, true)
            rt, rok := exprType(as.RHS, false)
            if lok && rok {
                subst := map[string]astpkg.TypeRef{}
                if !unify(lt, rt, subst) {
                    d := diag.Diagnostic{Level: diag.Error, Code: "E_ASSIGN_TYPE_MISMATCH", Message: "assignment type mismatch: " + typeStr(lt) + " != " + typeStr(rt)}
                    d.Pos = &srcset.Position{Line: as.Pos.Line, Column: as.Pos.Column, Offset: as.Pos.Offset}
                    diags = append(diags, d)
                }
            }
            return
        }
        switch v := s.(type) {
        case astpkg.VarDeclStmt:
            // Handle var decls: var name [Type] [= init]
            if v.Type.Name != "" && v.Init != nil {
                rt, rok := exprType(v.Init, false)
                if rok {
                    subst := map[string]astpkg.TypeRef{}
                    if !unify(v.Type, rt, subst) {
                        d := diag.Diagnostic{Level: diag.Error, Code: "E_ASSIGN_TYPE_MISMATCH", Message: "var init type mismatch: " + typeStr(v.Type) + " != " + typeStr(rt)}
                        d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
                        diags = append(diags, d)
                    }
                }
                env[v.Name] = v.Type
            } else if v.Type.Name != "" && v.Init == nil {
                env[v.Name] = v.Type
            } else if v.Type.Name == "" && v.Init != nil {
                if rt, rok := exprType(v.Init, false); rok {
                    env[v.Name] = rt
                } else {
                    d := diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_UNINFERRED", Message: "cannot infer variable type from initializer"}
                    d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
                    diags = append(diags, d)
                }
            } else {
                // no type and no init
                d := diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_UNINFERRED", Message: "variable declaration missing type and initializer"}
                d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
                diags = append(diags, d)
            }
        case astpkg.ExprStmt:
            // type derivation for expression statements (ensures BinaryExpr checks run)
            _, _ = exprType(v.X, false)
        case astpkg.ReturnStmt:
            // validate return types
            if len(fd.Result) == 0 {
                if len(v.Results) > 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "function has no return values"})
                }
                return
            }
            if len(fd.Result) != len(v.Results) {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return value count does not match function result arity"})
                return
            }
            // unify each returned expression against declared result
            for i, rexpr := range v.Results {
                rt, rok := exprType(rexpr, false)
                if !rok {
                    continue
                }
                want := fd.Result[i]
                subst := map[string]astpkg.TypeRef{}
                if !unify(want, rt, subst) {
                    d := diag.Diagnostic{Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return type mismatch: got " + typeStr(rt) + ", want " + typeStr(want)}
                    d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
                    diags = append(diags, d)
                    continue
                }
                // if any substitution binds to a type variable (not concrete), report uninferred
                for _, b := range subst {
                    if len(b.Name) == 1 && b.Name[0] >= 'A' && b.Name[0] <= 'Z' {
                        d := diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_UNINFERRED", Message: "return type contains uninferred type variables"}
                        d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
                        diags = append(diags, d)
                        break
                    }
                }
            }
        case astpkg.BlockStmt:
            for _, ss := range v.Stmts {
                walkStmt(ss)
            }
        }
    }
    for _, s := range fd.BodyStmts {
        walkStmt(s)
    }
    return diags
}
