package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "strings"
)

// analyzeRAII enforces minimal ownership/RAII rules for Owned<T> parameters:
// - E_RAII_OWNED_NOT_RELEASED: Owned param must be released or transferred.
// - E_RAII_DOUBLE_RELEASE: multiple releases/transfers for same variable.
// - E_RAII_USE_AFTER_RELEASE: use of variable after release/transfer.
// Release/transfer detection:
// - Calls to functions whose corresponding param type is Owned<…>.
// - Calls to known releasers: release(x), drop(x), free(x), dispose(x) or x.Close()/x.Release()/x.Free()/x.Dispose().
func analyzeRAII(fd astpkg.FuncDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Body) == 0 {
        return diags
    }
    if len(fd.BodyStmts) > 0 {
        return analyzeRAIIFromAST(fd, funcs)
    }
    // Token-based fallback analysis
    // collect Owned<T> parameters by name
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
    usedAfter := map[string]bool{}
    // helper: process function call at tokens[i] being callee IDENT or receiver IDENT '.' method
    toks := fd.Body
    isReleaser := func(name string) bool {
        switch strings.ToLower(name) {
        case "release", "drop", "free", "dispose":
            return true
        }
        return false
    }
    // parse call args starting at index of '('; returns list of top-level identifier args and end index of ')'
    parseArgs := func(start int) ([]string, int) {
        args := []string{}
        depth := 0
        curIdent := ""
        end := start
        for i := start; i < len(toks); i++ {
            end = i
            tk := toks[i]
            if tk.Kind == tok.LPAREN {
                depth++
                continue
            }
            if tk.Kind == tok.RPAREN {
                depth--
                if depth == 0 {
                    break
                }
                continue
            }
            if depth == 1 { // top-level within call
                if tk.Kind == tok.IDENT {
                    curIdent = tk.Lexeme
                } else {
                    if curIdent != "" {
                        args = append(args, curIdent)
                        curIdent = ""
                    }
                }
                if tk.Kind == tok.COMMA {
                    if curIdent != "" {
                        args = append(args, curIdent)
                        curIdent = ""
                    }
                }
            }
        }
        if curIdent != "" {
            args = append(args, curIdent)
        }
        return args, end
    }
    // map for known function signatures
    // scan tokens
    for i := 0; i < len(toks); i++ {
        t := toks[i]
        // use-after-release detection (simple): any occurrence of owned ident after release
        if t.Kind == tok.IDENT && owned[t.Lexeme] && released[t.Lexeme] {
            if !usedAfter[t.Lexeme] {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
                usedAfter[t.Lexeme] = true
            }
        }
        // function call: IDENT '(' ... ')'
        if t.Kind == tok.IDENT && i+1 < len(toks) && toks[i+1].Kind == tok.LPAREN {
            callee := t.Lexeme
            args, end := parseArgs(i + 1)
            // if callee is known function, transfer ownership for args matching Owned params
            if fd2, ok := funcs[callee]; ok {
                // detect if callee accepts any Owned parameter (position-agnostic fallback)
                calleeHasOwned := false
                for _, p := range fd2.Params {
                    if strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
                        calleeHasOwned = true
                        break
                    }
                }
                matched := false
                for idx, a := range args {
                    if !owned[a] {
                        continue
                    }
                    // precise positional check
                    transferred := false
                    if idx < len(fd2.Params) {
                        pt := fd2.Params[idx].Type
                        if strings.ToLower(pt.Name) == "owned" && len(pt.Args) == 1 {
                            transferred = true
                        }
                    }
                    if !transferred && calleeHasOwned {
                        transferred = true
                    }
                    if transferred {
                        if released[a] {
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
                        }
                        released[a] = true
                        matched = true
                    }
                }
                if !matched && calleeHasOwned {
                    // heuristic fallback: release any single owned param if only one exists
                    count := 0
                    last := ""
                    for name := range owned {
                        count++
                        last = name
                    }
                    if count == 1 {
                        if released[last] {
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
                        }
                        released[last] = true
                    }
                }
            }
            // known releaser by name
            if isReleaser(callee) {
                if len(args) == 0 {
                    // conservative fallback if only one owned param exists
                    count := 0
                    last := ""
                    for name := range owned {
                        count++
                        last = name
                    }
                    if count == 1 {
                        if released[last] {
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                        }
                        released[last] = true
                    }
                }
                for _, a := range args {
                    if owned[a] {
                        if released[a] {
                            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                        }
                        released[a] = true
                    }
                }
            }
            i = end
            continue
        }
        // method call: IDENT '.' IDENT '(' ... ')'
        if t.Kind == tok.IDENT && i+3 < len(toks) && toks[i+1].Kind == tok.DOT && toks[i+2].Kind == tok.IDENT && toks[i+3].Kind == tok.LPAREN {
            recv := t.Lexeme
            mth := toks[i+2].Lexeme
            // parse args to matching ')'
            _, end := parseArgs(i + 3)
            if owned[recv] {
                switch strings.ToLower(mth) {
                case "close", "release", "free", "dispose":
                    if released[recv] {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
                    }
                    released[recv] = true
                }
            }
            i = end
            continue
        }
    }
    // end-of-function: any owned param not released/transferred -> diagnostic
    if !isSinkFunction(fd) {
        for name := range owned {
            if !released[name] {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_OWNED_NOT_RELEASED", Message: "owned value not released or transferred before function end"})
            }
        }
    }
    return diags
}

// isSinkFunction heuristically identifies helper functions that consume Owned<T>
// parameters and are not subject to RAII not-released enforcement. Criteria:
// - Function is not a worker signature; and
// - Has at least one parameter of type Owned<…>; and
// - Has exactly one parameter (common sink pattern).
func isSinkFunction(fd astpkg.FuncDecl) bool {
    if isWorkerSignature(fd) {
        return false
    }
    hasOwned := false
    for _, p := range fd.Params {
        if strings.EqualFold(p.Type.Name, "owned") && len(p.Type.Args) == 1 {
            hasOwned = true
            break
        }
    }
    if !hasOwned {
        return false
    }
    if len(fd.Params) == 1 {
        return true
    }
    return false
}

