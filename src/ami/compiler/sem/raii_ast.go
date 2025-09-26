package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

func analyzeRAIIFromAST(fd astpkg.FuncDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // owned params
    owned := map[string]bool{}
    for _, p := range fd.Params {
        if p.Name != "" && strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
            owned[p.Name] = true
        }
    }
    if len(owned) == 0 {
        return diags
    }
    released := map[string]bool{}
    deferred := map[string]bool{}
    usedAfter := map[string]bool{}
    // helpers
    isReleaser := func(name string) bool {
        switch strings.ToLower(name) {
        case "release", "drop", "free", "dispose":
            return true
        }
        return false
    }
    var walkExpr func(astpkg.Expr)
    walkExpr = func(e astpkg.Expr) {
        switch v := e.(type) {
        case astpkg.Ident:
            if owned[v.Name] && released[v.Name] && !usedAfter[v.Name] {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
                usedAfter[v.Name] = true
            }
        case astpkg.UnaryExpr:
            walkExpr(v.X)
        case astpkg.SelectorExpr:
            // receiver use counts as use-after
            if id, ok := v.X.(astpkg.Ident); ok {
                if owned[id.Name] {
                    if strings.EqualFold(v.Sel, "close") || strings.EqualFold(v.Sel, "release") || strings.EqualFold(v.Sel, "free") || strings.EqualFold(v.Sel, "dispose") {
                        if released[id.Name] || deferred[id.Name] {
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                        }
                        released[id.Name] = true
                    } else if released[id.Name] && !usedAfter[id.Name] {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
                        usedAfter[id.Name] = true
                    }
                }
            }
        case astpkg.CallExpr:
            // function/method calls
            switch f := v.Fun.(type) {
            case astpkg.Ident:
                name := f.Name
                // releaser by name
                if isReleaser(name) {
                    mark := false
                    for _, a := range v.Args {
                        if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
                            if released[id.Name] || deferred[id.Name] {
                                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                            }
                            released[id.Name] = true
                            mark = true
                        }
                    }
                    if !mark {
                        // conservative: release single owned param if only one
                        if len(owned) == 1 {
                            for k := range owned {
                                released[k] = true
                            }
                        }
                    }
                }
                // Always visit args to catch nested calls (e.g., mutate(release(x)))
                for _, a := range v.Args {
                    walkExpr(a)
                }
                // transfer via owned parameter position
                if callee, ok := funcs[name]; ok {
                    calleeHasOwned := false
                    for _, p := range callee.Params {
                        if strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
                            calleeHasOwned = true
                            break
                        }
                    }
                    matched := false
                    for idx, a := range v.Args {
                        if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
                            if idx < len(callee.Params) {
                                pt := callee.Params[idx].Type
                                if strings.ToLower(pt.Name) == "owned" && len(pt.Args) == 1 {
                                    released[id.Name] = true
                                    matched = true
                                }
                            }
                        }
                    }
                    if !matched && calleeHasOwned && len(owned) == 1 {
                        for k := range owned {
                            released[k] = true
                        }
                    }
                }
            case astpkg.SelectorExpr:
                // Method call: receiver.method(args)
                // Treat known releaser methods as release operations.
                if id, ok := f.X.(astpkg.Ident); ok && owned[id.Name] {
                    if strings.EqualFold(f.Sel, "close") || strings.EqualFold(f.Sel, "release") || strings.EqualFold(f.Sel, "free") || strings.EqualFold(f.Sel, "dispose") {
                        if released[id.Name] || deferred[id.Name] {
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                        }
                        deferred[id.Name] = true
                    } else if released[id.Name] && !usedAfter[id.Name] {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
                        usedAfter[id.Name] = true
                    }
                }
                // Visit args for use-after detection in their subtrees
                for _, a := range v.Args {
                    walkExpr(a)
                }
            default:
                for _, a := range v.Args {
                    walkExpr(a)
                }
            }
        }
    }
    var walkStmt func(astpkg.Stmt)
    walkStmt = func(s astpkg.Stmt) {
        switch v := s.(type) {
        case astpkg.AssignStmt:
            walkExpr(v.LHS)
            walkExpr(v.RHS)
        case astpkg.ExprStmt:
            walkExpr(v.X)
        case astpkg.DeferStmt:
            // analyze deferred call specially: schedule release/transfer at end
            if ce, ok := v.X.(astpkg.CallExpr); ok {
                switch f := ce.Fun.(type) {
                case astpkg.Ident:
                    name := f.Name
                    // known releaser by name
                    if isReleaser(name) {
                        mark := false
                        for _, a := range ce.Args {
                            if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
                                if released[id.Name] || deferred[id.Name] {
                                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                                }
                                deferred[id.Name] = true
                                mark = true
                            }
                        }
                        if !mark {
                            if len(owned) == 1 {
                                for k := range owned {
                                    if released[k] {
                                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                                    }
                                    deferred[k] = true
                                }
                            }
                        }
                    }
                    // Transfer semantics for known functions with Owned params: treat as release at end
                    if callee, ok := funcs[name]; ok {
                        calleeHasOwned := false
                        for _, p := range callee.Params {
                            if strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
                                calleeHasOwned = true
                                break
                            }
                        }
                        matched := false
                        for idx, a := range ce.Args {
                            if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
                                if idx < len(callee.Params) {
                                    pt := callee.Params[idx].Type
                                    if strings.ToLower(pt.Name) == "owned" && len(pt.Args) == 1 {
                                        if released[id.Name] || deferred[id.Name] {
                                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
                                        }
                                        deferred[id.Name] = true
                                        matched = true
                                    }
                                }
                            }
                        }
                        if !matched && calleeHasOwned && len(owned) == 1 {
                            for k := range owned {
                                if released[k] || deferred[k] {
                                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
                                }
                                deferred[k] = true
                            }
                        }
                    }
                case astpkg.SelectorExpr:
                    // receiver.method(...)
                    if id, ok := f.X.(astpkg.Ident); ok && owned[id.Name] {
                        if strings.EqualFold(f.Sel, "close") || strings.EqualFold(f.Sel, "release") || strings.EqualFold(f.Sel, "free") || strings.EqualFold(f.Sel, "dispose") {
                            if released[id.Name] || deferred[id.Name] {
                                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                            }
                            deferred[id.Name] = true
                        }
                    }
                default:
                    // nothing
                }
            }
        case astpkg.MutBlockStmt:
            for _, ss := range v.Body.Stmts {
                walkStmt(ss)
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
    if !isSinkFunction(fd) {
        for name := range owned {
            if !(released[name] || deferred[name]) {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_OWNED_NOT_RELEASED", Message: "owned value not released or transferred before function end"})
            }
        }
    }
    return diags
}

