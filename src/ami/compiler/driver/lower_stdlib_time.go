package driver

import (
    "strings"
    "hash/fnv"
    "strconv"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// handlerTokenImmediate returns an immediate ID ("#<num>") representing a deterministic
// token for a handler expression. Supports identifiers and selector expressions (alias-qualified).
// For any other form (future: lambdas), falls back to a stable position-based token.
func handlerTokenImmediate(e ast.Expr) (string, bool) {
    name := ""
    switch v := e.(type) {
    case *ast.IdentExpr:
        name = v.Name
    case *ast.SelectorExpr:
        // Flatten selector chain to text, e.g., pkg.Func or a.b.c
        name = selectorText(v)
    }
    if name != "" {
        h := fnv.New64a(); _, _ = h.Write([]byte(name))
        tok := int64(h.Sum64())
        return "#" + strconv.FormatInt(tok, 10), true
    }
    // Fallback: use position offset to create a deterministic token within the file
    off := exprOffset(e)
    if off >= 0 {
        h := fnv.New64a(); _, _ = h.Write([]byte("anon@" + strconv.Itoa(off)))
        tok := int64(h.Sum64())
        return "#" + strconv.FormatInt(tok, 10), true
    }
    return "", false
}

// handlerTokenValue computes the deterministic token value for a handler expression.
func handlerTokenValue(e ast.Expr) (int64, bool) {
    name := ""
    switch v := e.(type) {
    case *ast.IdentExpr:
        name = v.Name
    case *ast.SelectorExpr:
        name = selectorText(v)
    }
    if name != "" {
        h := fnv.New64a(); _, _ = h.Write([]byte(name))
        return int64(h.Sum64()), true
    }
    off := exprOffset(e)
    if off >= 0 {
        h := fnv.New64a(); _, _ = h.Write([]byte("anon@" + strconv.Itoa(off)))
        return int64(h.Sum64()), true
    }
    return 0, false
}

func selectorText(s *ast.SelectorExpr) string {
    if s == nil { return "" }
    left := ""
    switch v := s.X.(type) {
    case *ast.IdentExpr:
        left = v.Name
    case *ast.SelectorExpr:
        left = selectorText(v)
    default:
        left = "?"
    }
    if left == "" { left = "?" }
    return left + "." + s.Sel
}

func exprOffset(e ast.Expr) int {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Pos.Offset
    case *ast.CallExpr:
        return v.Pos.Offset
    case *ast.SelectorExpr:
        return v.Pos.Offset
    case *ast.StringLit:
        return v.Pos.Offset
    case *ast.NumberLit:
        return v.Pos.Offset
    default:
        return -1
    }
}

// lowerStdlibCall recognizes AMI stdlib calls and lowers them to runtime intrinsics
// or optimized IR forms. It returns (expr, true) when handled.
func lowerStdlibCall(st *lowerState, c *ast.CallExpr) (ir.Expr, bool) {
    if c == nil { return ir.Expr{}, false }
    name := c.Name
    // Normalize alias-qualified call by suffix when possible
    // Supported time intrinsics: time.Sleep(d)
    // Supported signal intrinsic: signal.Register(sig, fn)
    if strings.HasSuffix(name, ".Register") || name == "signal.Register" {
        var args []ir.Value
        // arg0: signal enum → i64 immediate token. Prefer selector mapping for stability.
        if len(c.Args) >= 1 {
            switch s := c.Args[0].(type) {
            case *ast.SelectorExpr:
                // Map a few common signals; otherwise, fallback to lowered expr coercion.
                var v int64
                switch s.Sel {
                case "SIGINT": v = 2
                case "SIGTERM": v = 15
                case "SIGHUP": v = 1
                case "SIGQUIT": v = 3
                default:
                    if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil {
                        args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"})
                    }
                }
                if v != 0 { args = append(args, ir.Value{ID: "#"+strconv.FormatInt(v, 10), Type: "int64"}) }
            default:
                if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil {
                    args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"})
                }
            }
        }
        // arg1: handler function reference → opaque handler token (i64) with deterministic hash of name
        if len(c.Args) >= 2 {
            if tokID, ok := handlerTokenImmediate(c.Args[1]); ok {
                args = append(args, ir.Value{ID: tokID, Type: "int64"})
            } else {
                if ex, ok := lowerExpr(st, c.Args[1]); ok && ex.Result != nil {
                    args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"})
                }
            }
        }
        return ir.Expr{Op: "call", Callee: "ami_rt_signal_register", Args: args}, true
    }
    if strings.HasSuffix(name, ".Sleep") || name == "time.Sleep" {
        // Lower to runtime sleep (milliseconds). Result is void.
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        return ir.Expr{Op: "call", Callee: "ami_rt_sleep_ms", Args: args}, true
    }
    if strings.HasSuffix(name, ".Now") || name == "time.Now" {
        // Result is a time handle (opaque i64); use AMI type "Time" for tracking
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "Time"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_now", Result: res}, true
    }
    if strings.HasSuffix(name, ".Add") || name == "time.Add" {
        // Args: (Time handle, d ms)
        var args []ir.Value
        if len(c.Args) >= 1 {
            if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil {
                // coerce time handle to int64 for runtime ABI
                args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"})
            }
        }
        if len(c.Args) >= 2 {
            if ex, ok := lowerExpr(st, c.Args[1]); ok && ex.Result != nil { args = append(args, *ex.Result) }
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Time"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_add", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".Delta") || name == "time.Delta" {
        var args []ir.Value
        for i := 0; i < len(c.Args) && i < 2; i++ {
            if ex, ok := lowerExpr(st, c.Args[i]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) }
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_delta", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".UnixNano") || name == "time.UnixNano" {
        var args []ir.Value
        if len(c.Args) >= 1 {
            if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) }
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_unix_nano", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".Unix") || name == "time.Unix" {
        var args []ir.Value
        if len(c.Args) >= 1 {
            if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) }
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_unix", Args: args, Result: res}, true
    }
    // Future signal handler primitives: Install, Token
    if strings.HasSuffix(name, ".Install") || name == "signal.Install" {
        // args: (fn any)
        if len(c.Args) >= 1 {
            // compute handler token (deterministic)
            tok, ok := handlerTokenValue(c.Args[0])
            var args []ir.Value
            if ok {
                args = append(args, ir.Value{ID: "#"+strconv.FormatInt(tok, 10), Type: "int64"})
            } else {
                // fallback: hash zero
                args = append(args, ir.Value{ID: "#0", Type: "int64"})
            }
            // function pointer symbol immediate for simple identifiers
            switch v := c.Args[0].(type) {
            case *ast.IdentExpr:
                if v.Name != "" {
                    args = append(args, ir.Value{ID: "#@"+v.Name, Type: "ptr"})
                } else {
                    args = append(args, ir.Value{ID: "#null", Type: "ptr"})
                }
            default:
                args = append(args, ir.Value{ID: "#null", Type: "ptr"})
            }
            return ir.Expr{Op: "call", Callee: "ami_rt_install_handler_thunk", Args: args}, true
        }
        return ir.Expr{Op: "call", Callee: "ami_rt_install_handler_thunk", Args: []ir.Value{{ID: "#0", Type: "int64"}, {ID: "#null", Type: "ptr"}}}, true
    }
    if strings.HasSuffix(name, ".Token") || name == "signal.Token" {
        if len(c.Args) >= 1 {
            tok, ok := handlerTokenValue(c.Args[0])
            id := st.newTemp()
            res := &ir.Value{ID: id, Type: "int64"}
            if ok {
                return ir.Expr{Op: "lit:" + strconv.FormatInt(tok, 10), Result: res}, true
            }
            return ir.Expr{Op: "lit:0", Result: res}, true
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "lit:0", Result: res}, true
    }
    return ir.Expr{}, false
}
