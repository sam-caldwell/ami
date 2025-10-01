package llvm

import (
    "fmt"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFunction converts an ir.Function into a textual LLVM function definition.
func lowerFunction(fn ir.Function) (string, error) {
    // Signature
    ret := "void"
    switch len(fn.Results) {
    case 0:
        ret = "void"
    case 1:
        // Enforce backend memory safety: no raw pointers in public ABI.
        if isUnsafePointerType(fn.Results[0].Type) {
            return "", fmt.Errorf("unsafe pointer type in result: %s", fn.Results[0].Type)
        }
        ret = abiType(fn.Results[0].Type)
    default:
        // Aggregate return type: {T0, T1, ...}
        var parts []string
        for _, r := range fn.Results {
            if isUnsafePointerType(r.Type) { return "", fmt.Errorf("unsafe pointer type in result: %s", r.Type) }
            parts = append(parts, abiType(r.Type))
        }
        ret = "{" + strings.Join(parts, ", ") + "}"
    }
    // Params
    var params []string
    for _, p := range fn.Params {
        if isUnsafePointerType(p.Type) {
            return "", fmt.Errorf("unsafe pointer type in param: %s", p.Type)
        }
        params = append(params, fmt.Sprintf("%s %%%s", abiType(p.Type), p.ID))
    }
    // Body
    var b strings.Builder
    fmt.Fprintf(&b, "define %s @%s(%s) {\n", ret, fn.Name, strings.Join(params, ", "))
    // Collect defers to emit before any return (simulate run-at-exit semantics)
    var defers []ir.Expr
    for i, blk := range fn.Blocks {
        // label
        name := blk.Name
        if name == "" { name = fmt.Sprintf("b%d", i) }
        fmt.Fprintf(&b, "%s:\n", name)
        for _, ins := range blk.Instr {
            switch v := ins.(type) {
            case ir.Var:
                b.WriteString(lowerVar(v))
            case ir.Assign:
                b.WriteString(lowerAssign(v))
            case ir.Expr:
                b.WriteString(lowerExpr(v))
            case ir.Phi:
                b.WriteString(lowerPhi(v))
            case ir.CondBr:
                // Conditional branch on boolean i1
                // Ensure condition is mapped as i1; assume upstream typing enforces this
                fmt.Fprintf(&b, "  br i1 %%%s, label %%%s, label %%%s\n", v.Cond.ID, v.TrueLabel, v.FalseLabel)
            case ir.Defer:
                // Schedule defer for emission before returns (LIFO order)
                defers = append(defers, v.Expr)
            case ir.Return:
                // Emit deferred expressions in reverse (LIFO) before return
                for i := len(defers)-1; i >= 0; i-- { b.WriteString(lowerExpr(defers[i])) }
                s, err := lowerReturn(v)
                if err != nil { return "", err }
                b.WriteString(s)
            case ir.Loop:
                fmt.Fprintf(&b, "  ; loop %s\n", v.Name)
            case ir.Goto:
                fmt.Fprintf(&b, "  br label %%%s\n", v.Label)
            case ir.SetPC:
                fmt.Fprintf(&b, "  ; setpc %d\n", v.PC)
            case ir.Dispatch:
                fmt.Fprintf(&b, "  ; dispatch %s\n", v.Label)
            case ir.PushFrame:
                fmt.Fprintf(&b, "  ; push_frame %s\n", v.Fn)
            case ir.PopFrame:
                fmt.Fprintf(&b, "  ; pop_frame\n")
            default:
                // Unknown instruction (future): keep deterministic comment
                b.WriteString("  ; instr\n")
            }
        }
    }
    // Ensure a terminator exists: synthesize a default return when missing
    if !strings.Contains(b.String(), "\n  ret ") {
        // Emit deferred expressions in reverse order at function end
        for i := len(defers)-1; i >= 0; i-- { b.WriteString(lowerExpr(defers[i])) }
        if ret == "void" {
            b.WriteString("  ret void\n")
        } else {
            // Zero initializer for known scalars; use 0 or null pointer
            zero := "0"
            if ret == "ptr" { zero = "null" }
            fmt.Fprintf(&b, "  ret %s %s\n", ret, zero)
        }
    }
    b.WriteString("}\n")
    return b.String(), nil
}
