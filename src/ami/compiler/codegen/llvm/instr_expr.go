package llvm

import (
    "fmt"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerExpr emits expression instructions; supports call in scaffold.
func lowerExpr(e ir.Expr) string {
    // Integer literal scaffold: op form "lit:<int>"
    if strings.HasPrefix(e.Op, "lit:") {
        if e.Result != nil && e.Result.ID != "" {
            val := strings.TrimPrefix(e.Op, "lit:")
            // Choose lowering based on requested result type.
            ty := mapType(e.Result.Type)
            if ty == "i1" { // boolean literal
                if val == "0" || strings.EqualFold(val, "false") {
                    return fmt.Sprintf("  %%%s = icmp ne i1 0, 1\n", e.Result.ID) // always false
                }
                return fmt.Sprintf("  %%%s = icmp eq i1 0, 0\n", e.Result.ID) // always true
            }
            if ty == "double" { // float literal (scaffold)
                if val == "0" || val == "0.0" { return fmt.Sprintf("  %%%s = fadd double 0.0, 0.0\n", e.Result.ID) }
                return fmt.Sprintf("  %%%s = fadd double 0.0, %s\n", e.Result.ID, val)
            }
            // default integer path
            return fmt.Sprintf("  %%%s = add i64 0, %s\n", e.Result.ID, val)
        }
    }
    if strings.EqualFold(e.Op, "call") {
        // Return type resolution
        // - Runtime helpers: use their true ABI regardless of whether result captured
        // - User functions: map via ABI, avoid raw ptr exposure when captured
        ret := "void"
        if strings.HasPrefix(e.Callee, "ami_rt_") {
            switch e.Callee {
            case "ami_rt_panic", "ami_rt_zeroize", "ami_rt_zeroize_owned":
                ret = "void"
            case "ami_rt_signal_register":
                ret = "void"
            case "ami_rt_owned_len":
                ret = "i64"
            case "ami_rt_string_len", "ami_rt_slice_len":
                ret = "i64"
            case "ami_rt_alloc", "ami_rt_owned_ptr", "ami_rt_owned_new":
                ret = "ptr"
            case "ami_rt_sleep_ms":
                ret = "void"
            case "ami_rt_time_now":
                ret = "i64"
            case "ami_rt_time_add":
                ret = "i64"
            case "ami_rt_time_delta":
                ret = "i64"
            case "ami_rt_time_unix":
                ret = "i64"
            case "ami_rt_time_unix_nano":
                ret = "i64"
            case "ami_rt_install_handler_thunk":
                ret = "void"
            case "ami_rt_get_handler_thunk":
                ret = "ptr"
            default:
                // fall back to result type when provided
                if e.Result != nil { ret = mapType(e.Result.Type) } else { ret = "void" }
            }
        } else if e.Result != nil {
            // Avoid exposing raw pointers at the language ABI boundary for user functions
            rt := mapType(e.Result.Type)
            if rt == "ptr" { rt = "i64" }
            ret = rt
        }
        // Args
        var args []string
        for _, a := range e.Args {
            ty := mapType(a.Type)
            if strings.HasPrefix(a.ID, "#@") {
                // Immediate function pointer symbol: ID "#@NAME" → ptr @NAME
                args = append(args, fmt.Sprintf("%s @%s", ty, strings.TrimPrefix(a.ID, "#@")))
                continue
            }
            if strings.HasPrefix(a.ID, "#") {
                // Immediate constant argument encoded as ID "#<literal>" (e.g., numbers, null)
                lit := strings.TrimPrefix(a.ID, "#")
                // special-case null pointer literal
                if lit == "null" && ty == "ptr" {
                    args = append(args, fmt.Sprintf("%s null", ty))
                } else {
                    args = append(args, fmt.Sprintf("%s %s", ty, lit))
                }
                continue
            }
            args = append(args, fmt.Sprintf("%s %%%s", ty, a.ID))
        }
        if e.Result != nil && e.Result.ID != "" {
            return fmt.Sprintf("  %%%s = call %s @%s(%s)\n", e.Result.ID, ret, e.Callee, strings.Join(args, ", "))
        }
        return fmt.Sprintf("  call %s @%s(%s)\n", ret, e.Callee, strings.Join(args, ", "))
    }
    // Basic arithmetic scaffold
    op := strings.ToLower(e.Op)
    switch op {
    case "select":
        // Retained for backward compatibility; prefer CFG + PHI now.
        if len(e.Args) >= 3 && e.Result != nil && e.Result.ID != "" {
            ty := mapType(e.Result.Type)
            if ty == "void" || ty == "" { ty = mapType(e.Args[1].Type) }
            return fmt.Sprintf("  %%%s = select i1 %%%s, %s %%%s, %s %%%s\n",
                e.Result.ID, e.Args[0].ID, ty, e.Args[1].ID, ty, e.Args[2].ID)
        }
        return "  ; expr select\n"
    case "load":
        // load <ty>, ptr %p
        if e.Result != nil && e.Result.ID != "" && len(e.Args) >= 1 {
            ty := mapType(e.Result.Type)
            if ty == "" || ty == "void" { ty = "i64" }
            return fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", e.Result.ID, ty, e.Args[0].ID)
        }
        return "  ; expr load\n"
    case "store":
        // store <ty> %v, ptr %p
        if len(e.Args) >= 2 {
            ty := mapType(e.Args[0].Type)
            if ty == "" || ty == "void" { ty = "i64" }
            return fmt.Sprintf("  store %s %%%s, ptr %%%s\n", ty, e.Args[0].ID, e.Args[1].ID)
        }
        return "  ; expr store\n"
    case "gep":
        // getelementptr i8, ptr %base, i64 %idx, ... (byte addressing for scaffold)
        if e.Result != nil && e.Result.ID != "" && len(e.Args) >= 2 {
            var idxs []string
            for i := 1; i < len(e.Args); i++ { idxs = append(idxs, fmt.Sprintf("i64 %%%s", e.Args[i].ID)) }
            return fmt.Sprintf("  %%%s = getelementptr i8, ptr %%%s, %s\n", e.Result.ID, e.Args[0].ID, strings.Join(idxs, ", "))
        }
        return "  ; expr gep\n"
    case "add", "sub", "mul", "div", "mod":
        // choose operation mnemonic by type (double → f*, else integer)
        ty := "i64"
        if e.Result != nil {
            mt := mapType(e.Result.Type)
            if mt == "double" { ty = "double" } else if mt != "ptr" { ty = mt }
        } else if len(e.Args) > 0 {
            mt := mapType(e.Args[0].Type)
            if mt == "double" { ty = "double" }
        }
        var mnem string
        switch op {
        case "add": mnem = map[string]string{"i64":"add","double":"fadd"}[ty]
        case "sub": mnem = map[string]string{"i64":"sub","double":"fsub"}[ty]
        case "mul": mnem = map[string]string{"i64":"mul","double":"fmul"}[ty]
        case "div":
            if ty == "double" { mnem = "fdiv" } else { mnem = "sdiv" }
        case "mod":
            if ty == "double" { return "  ; expr mod-unsupported-double\n" }
            mnem = "srem"
        }
        if e.Result != nil && e.Result.ID != "" && len(e.Args) >= 2 {
            return fmt.Sprintf("  %%%s = %s %s %%%s, %%%s\n", e.Result.ID, mnem, ty, e.Args[0].ID, e.Args[1].ID)
        }
        if len(e.Args) >= 2 {
            return fmt.Sprintf("  %s %s %%%s, %%%s\n", mnem, ty, e.Args[0].ID, e.Args[1].ID)
        }
    case "eq", "ne", "lt", "le", "gt", "ge":
        if len(e.Args) >= 2 {
            // choose cmp mnemonic (int vs double)
            ty := "i64"
            if len(e.Args) > 0 {
                mt := mapType(e.Args[0].Type)
                if mt == "double" { ty = "double" }
            }
            if e.Result != nil && e.Result.ID != "" {
                if ty == "double" {
                    pred := map[string]string{"eq": "oeq", "ne": "one", "lt": "olt", "le": "ole", "gt": "ogt", "ge": "oge"}[op]
                    return fmt.Sprintf("  %%%s = fcmp %s %s %%%s, %%%s\n", e.Result.ID, pred, ty, e.Args[0].ID, e.Args[1].ID)
                }
                pred := map[string]string{"eq": "eq", "ne": "ne", "lt": "slt", "le": "sle", "gt": "sgt", "ge": "sge"}[op]
                return fmt.Sprintf("  %%%s = icmp %s %s %%%s, %%%s\n", e.Result.ID, pred, ty, e.Args[0].ID, e.Args[1].ID)
            }
        }
    case "and", "or":
        if len(e.Args) >= 2 {
            // Only boolean logical forms here; bitwise handled via band/bor.
            ty := "i1"
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt != "i1" { return "  ; expr logic-nonbool\n" }
            }
            mnem := op // "and" or "or"
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = %s %s %%%s, %%%s\n", e.Result.ID, mnem, ty, e.Args[0].ID, e.Args[1].ID)
            }
            return fmt.Sprintf("  %s %s %%%s, %%%s\n", mnem, ty, e.Args[0].ID, e.Args[1].ID)
        }
    case "band":
        if len(e.Args) >= 2 {
            // Integer bitwise and (never float/ptr)
            ty := "i64"
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt == "double" { return "  ; expr band-float\n" }
                if mt != "ptr" && mt != "" { ty = mt }
            } else if len(e.Args) > 0 {
                mt := mapType(e.Args[0].Type)
                if mt != "ptr" && mt != "double" && mt != "" { ty = mt }
            }
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = and %s %%%s, %%%s\n", e.Result.ID, ty, e.Args[0].ID, e.Args[1].ID)
            }
            return fmt.Sprintf("  and %s %%%s, %%%s\n", ty, e.Args[0].ID, e.Args[1].ID)
        }
    case "bor":
        if len(e.Args) >= 2 {
            ty := "i64"
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt == "double" { return "  ; expr bor-float\n" }
                if mt != "ptr" && mt != "" { ty = mt }
            } else if len(e.Args) > 0 {
                mt := mapType(e.Args[0].Type)
                if mt != "ptr" && mt != "double" && mt != "" { ty = mt }
            }
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = or %s %%%s, %%%s\n", e.Result.ID, ty, e.Args[0].ID, e.Args[1].ID)
            }
            return fmt.Sprintf("  or %s %%%s, %%%s\n", ty, e.Args[0].ID, e.Args[1].ID)
        }
    case "not":
        if len(e.Args) >= 1 && e.Result != nil && e.Result.ID != "" {
            // logical not for i1: xor with true
            return fmt.Sprintf("  %%%s = xor i1 %%%s, true\n", e.Result.ID, e.Args[0].ID)
        }
    case "xor":
        if len(e.Args) >= 2 {
            // integer xor (or boolean xor as i1)
            ty := "i64"
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt == "i1" { ty = "i1" } else if mt != "ptr" { ty = mt }
            }
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = xor %s %%%s, %%%s\n", e.Result.ID, ty, e.Args[0].ID, e.Args[1].ID)
            }
            return fmt.Sprintf("  xor %s %%%s, %%%s\n", ty, e.Args[0].ID, e.Args[1].ID)
        }
    case "shl":
        if len(e.Args) >= 2 {
            ty := "i64"
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt != "ptr" && mt != "double" { ty = mt }
            }
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = shl %s %%%s, %%%s\n", e.Result.ID, ty, e.Args[0].ID, e.Args[1].ID)
            }
            return fmt.Sprintf("  shl %s %%%s, %%%s\n", ty, e.Args[0].ID, e.Args[1].ID)
        }
    case "shr":
        if len(e.Args) >= 2 {
            ty := "i64"
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt != "ptr" && mt != "double" { ty = mt }
            }
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = ashr %s %%%s, %%%s\n", e.Result.ID, ty, e.Args[0].ID, e.Args[1].ID)
            }
            return fmt.Sprintf("  ashr %s %%%s, %%%s\n", ty, e.Args[0].ID, e.Args[1].ID)
        }
    case "neg":
        if len(e.Args) >= 1 {
            // unary negation of a value
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt == "double" {
                    return fmt.Sprintf("  %%%s = fsub double 0.0, %%%s\n", e.Result.ID, e.Args[0].ID)
                }
                // integer
                ty := mt
                if ty == "ptr" || ty == "" { ty = "i64" }
                return fmt.Sprintf("  %%%s = sub %s 0, %%%s\n", e.Result.ID, ty, e.Args[0].ID)
            }
        }
    case "bnot":
        if len(e.Args) >= 1 {
            // bitwise not via xor with -1
            ty := "i64"
            if e.Result != nil {
                mt := mapType(e.Result.Type)
                if mt != "ptr" && mt != "double" { ty = mt }
            }
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = xor %s %%%s, -1\n", e.Result.ID, ty, e.Args[0].ID)
            }
        }
    }
    return fmt.Sprintf("  ; expr %s\n", e.Op)
}
