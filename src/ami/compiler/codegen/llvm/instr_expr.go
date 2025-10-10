package llvm

import (
    "fmt"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
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
    // Field projection: op form "field.<path>"; define a result of appropriate type.
    if strings.HasPrefix(e.Op, "field.") {
        // Use base argument type to compute layout offsets when possible
        if e.Result != nil && e.Result.ID != "" && len(e.Args) >= 1 {
            base := e.Args[0]
            baseTy := base.Type
            path := strings.TrimPrefix(e.Op, "field.")
            if t, err := types.Parse(baseTy); err == nil {
                if offSlots, _, ok := fieldOffsetSlots(t, path); ok {
                    off := offSlots * 8
                    // Compute GEP to the field slot
                    gid := "gep_" + e.Result.ID
                    var b strings.Builder
                    fmt.Fprintf(&b, "  %%%s = getelementptr i8, ptr %%%s, i64 %d\n", gid, base.ID, off)
                    // Load leaf value typed by result
                    rty := mapType(e.Result.Type)
                    if rty == "" || rty == "void" { rty = "i64" }
                    fmt.Fprintf(&b, "  %%%s = load %s, ptr %%%s\n", e.Result.ID, rty, gid)
                    return b.String()
                }
            }
            // Fallback when layout unknown: produce safe zero/empty values
            ty := mapType(e.Result.Type)
            switch ty {
            case "i1":
                return fmt.Sprintf("  %%%s = icmp ne i1 0, 1\n", e.Result.ID)
            case "i64":
                return fmt.Sprintf("  %%%s = add i64 0, 0\n", e.Result.ID)
            case "double":
                return fmt.Sprintf("  %%%s = fadd double 0.0, 0.0\n", e.Result.ID)
            case "ptr":
                return fmt.Sprintf("  %%%s = getelementptr i8, ptr null, i64 0\n", e.Result.ID)
            default:
                return fmt.Sprintf("  %%%s = add i64 0, 0\n", e.Result.ID)
            }
        }
        return "  ; expr field.get\n"
    }
    // Event payload extraction: op form "event.payload"; result type determines bridge
    if strings.EqualFold(e.Op, "event.payload") {
        if e.Result != nil && e.Result.ID != "" && len(e.Args) >= 1 {
            // Only numeric primitives supported for now; int -> i64 via runtime helper
            ty := mapType(e.Result.Type)
            ev := e.Args[0]
            if ty == "i64" {
                return fmt.Sprintf("  %%%s = call i64 @ami_rt_event_payload_to_i64(ptr %%%s)\n", e.Result.ID, ev.ID)
            }
            if ty == "double" {
                // not yet supported: fallback zero
                return fmt.Sprintf("  %%%s = fadd double 0.0, 0.0\n", e.Result.ID)
            }
            if ty == "i1" {
                return fmt.Sprintf("  %%%s = icmp eq i1 0, 0\n", e.Result.ID)
            }
            // default zero int
            return fmt.Sprintf("  %%%s = add i64 0, 0\n", e.Result.ID)
        }
        return "  ; expr event.payload\n"
    }
    if strings.EqualFold(e.Op, "call") {
        // Return type resolution
        // - Runtime helpers: use their true ABI regardless of whether result captured
        // - User functions: map via ABI, avoid raw ptr exposure when captured
        ret := "void"
        callee := e.Callee
        intrinsicMapped := false
        // Map standard math calls to LLVM intrinsics when present
        if strings.HasPrefix(callee, "math.") {
            switch callee {
            case "math.FMA": callee = "llvm.fma.f64"; ret = "double"
            case "math.Erf": callee = "llvm.erf.f64"; ret = "double"
            case "math.Erfc": callee = "llvm.erfc.f64"; ret = "double"
            case "math.Abs": callee = "llvm.fabs.f64"; ret = "double"
            case "math.Max": callee = "llvm.maxnum.f64"; ret = "double"
            case "math.Min": callee = "llvm.minnum.f64"; ret = "double"
            case "math.Ceil": callee = "llvm.ceil.f64"; ret = "double"
            case "math.Floor": callee = "llvm.floor.f64"; ret = "double"
            case "math.Trunc": callee = "llvm.trunc.f64"; ret = "double"
            case "math.Round": callee = "llvm.round.f64"; ret = "double"
            case "math.RoundToEven": callee = "llvm.roundeven.f64"; ret = "double"
            case "math.Exp": callee = "llvm.exp.f64"; ret = "double"
            case "math.Expm1": callee = "llvm.expm1.f64"; ret = "double"
            case "math.Exp2": callee = "llvm.exp2.f64"; ret = "double"
            case "math.Log1p": callee = "llvm.log1p.f64"; ret = "double"
            case "math.Log": callee = "llvm.log.f64"; ret = "double"
            case "math.Log2": callee = "llvm.log2.f64"; ret = "double"
            case "math.Log10": callee = "llvm.log10.f64"; ret = "double"
            case "math.Sqrt": callee = "llvm.sqrt.f64"; ret = "double"
            case "math.Pow": callee = "llvm.pow.f64"; ret = "double"
            case "math.Sin": callee = "llvm.sin.f64"; ret = "double"
            case "math.Cos": callee = "llvm.cos.f64"; ret = "double"
            case "math.Tan": callee = "llvm.tan.f64"; ret = "double"
            case "math.Asin": callee = "llvm.asin.f64"; ret = "double"
            case "math.Acos": callee = "llvm.acos.f64"; ret = "double"
            case "math.Atan": callee = "llvm.atan.f64"; ret = "double"
            case "math.Atan2": callee = "llvm.atan2.f64"; ret = "double"
            case "math.Sinh": callee = "llvm.sinh.f64"; ret = "double"
            case "math.Cosh": callee = "llvm.cosh.f64"; ret = "double"
            case "math.Tanh": callee = "llvm.tanh.f64"; ret = "double"
            case "math.Copysign": callee = "llvm.copysign.f64"; ret = "double"
            case "math.Nextafter": callee = "llvm.nextafter.f64"; ret = "double"
            case "math.Ldexp": callee = "llvm.ldexp.f64"; ret = "double"; intrinsicMapped = true
            case "math.Logb": callee = "ami_rt_math_logb"
            case "math.Ilogb": callee = "ami_rt_math_ilogb"
            }
        }
        // Precompute multi-result aggregate type when provided
        aggRet := ""
        if len(e.Results) > 1 {
            var parts []string
            for _, r := range e.Results { parts = append(parts, abiType(r.Type)) }
            aggRet = "{ " + strings.Join(parts, ", ") + " }"
        }
        if strings.HasPrefix(callee, "ami_rt_") {
            switch callee {
            case "ami_rt_panic", "ami_rt_zeroize", "ami_rt_zeroize_owned":
                ret = "void"
            case "ami_rt_signal_register":
                ret = "void"
            case "ami_rt_owned_len":
                ret = "i64"
            case "ami_rt_math_logb":
                ret = "double"
            case "ami_rt_math_ilogb":
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
            case "ami_rt_gpu_has":
                ret = "i1"
            case "ami_rt_cuda_devices", "ami_rt_opencl_platforms", "ami_rt_opencl_devices":
                ret = "ptr"
            default:
                // fall back to result type when provided
                if len(e.Results) > 1 {
                    // Multi-result runtime calls return an aggregate
                    var parts []string
                    for _, r := range e.Results { parts = append(parts, abiType(r.Type)) }
                    ret = "{ " + strings.Join(parts, ", ") + " }"
                } else if e.Result != nil {
                    ret = mapType(e.Result.Type)
                } else {
                    ret = "void"
                }
            }
        } else if len(e.Results) > 1 {
            ret = aggRet
        } else if e.Result != nil && !intrinsicMapped {
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
        if len(e.Results) > 1 {
            // Aggregate call then extract results
            aggID := "call_tup_" + e.Results[0].ID
            var sb strings.Builder
            fmt.Fprintf(&sb, "  %%%s = call %s @%s(%s)\n", aggID, ret, callee, strings.Join(args, ", "))
            for i, r := range e.Results {
                fmt.Fprintf(&sb, "  %%%s = extractvalue %s %%%s, %d\n", r.ID, ret, aggID, i)
            }
            return sb.String()
        }
        if e.Result != nil && e.Result.ID != "" {
            // Force intrinsic return types to double for known math intrinsics
            if strings.HasPrefix(callee, "llvm.") { ret = "double" }
            return fmt.Sprintf("  %%%s = call %s @%s(%s)\n", e.Result.ID, ret, callee, strings.Join(args, ", "))
        }
        if strings.HasPrefix(callee, "llvm.") { ret = "double" }
        return fmt.Sprintf("  call %s @%s(%s)\n", ret, callee, strings.Join(args, ", "))
    }
    // Basic arithmetic scaffold
    op := strings.ToLower(e.Op)
    switch op {
    case "frem":
        if len(e.Args) >= 2 {
            ty := "double"
            if e.Result != nil && e.Result.Type != "" { ty = mapType(e.Result.Type) }
            if e.Result != nil && e.Result.ID != "" {
                return fmt.Sprintf("  %%%s = frem %s %%%s, %%%s\n", e.Result.ID, ty, e.Args[0].ID, e.Args[1].ID)
            }
            return fmt.Sprintf("  frem %s %%%s, %%%s\n", ty, e.Args[0].ID, e.Args[1].ID)
        }
        return "  ; expr frem\n"
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
            if ty == "double" {
                if e.Result != nil && len(e.Args) >= 2 {
                    return fmt.Sprintf("  %%%s = frem double %%%s, %%%s\n", e.Result.ID, e.Args[0].ID, e.Args[1].ID)
                }
                return "  ; expr frem-double\n"
            }
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
