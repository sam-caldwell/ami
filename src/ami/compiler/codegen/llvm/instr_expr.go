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
            // Only lower integer literals here; floating is deferred.
            return fmt.Sprintf("  %%%s = add i64 0, %s\n", e.Result.ID, val)
        }
    }
    if strings.EqualFold(e.Op, "call") {
        // Return type (if any)
        ret := "void"
        if e.Result != nil { ret = mapType(e.Result.Type) }
        // Args
        var args []string
        for _, a := range e.Args {
            args = append(args, fmt.Sprintf("%s %%%s", mapType(a.Type), a.ID))
        }
        if e.Result != nil && e.Result.ID != "" {
            return fmt.Sprintf("  %%%s = call %s @%s(%s)\n", e.Result.ID, ret, e.Callee, strings.Join(args, ", "))
        }
        return fmt.Sprintf("  call %s @%s(%s)\n", ret, e.Callee, strings.Join(args, ", "))
    }
    // Basic arithmetic scaffold
    op := strings.ToLower(e.Op)
    switch op {
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
            // only defined for boolean (i1) in this scaffold
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
