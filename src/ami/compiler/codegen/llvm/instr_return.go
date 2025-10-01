package llvm

import (
    "fmt"
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerReturn emits a return instruction; single-value returns only in scaffold.
func lowerReturn(r ir.Return) (string, error) {
    switch len(r.Values) {
    case 0:
        return "  ret void\n", nil
    case 1:
        v := r.Values[0]
        return fmt.Sprintf("  ret %s %%%s\n", mapType(v.Type), v.ID), nil
    default:
        // Build aggregate {T0,T1,...} using insertvalue chain
        var parts []string
        for _, v := range r.Values { parts = append(parts, mapType(v.Type)) }
        aggTy := "{" + strings.Join(parts, ", ") + "}"
        // Start with undef
        var b strings.Builder
        name := fmt.Sprintf("ret_agg_%s", r.Values[0].ID)
        fmt.Fprintf(&b, "  %%%s0 = insertvalue %s undef, %s %%%s, 0\n", name, aggTy, mapType(r.Values[0].Type), r.Values[0].ID)
        prev := fmt.Sprintf("%%%s0", name)
        for i := 1; i < len(r.Values); i++ {
            cur := fmt.Sprintf("%%%s%d", name, i)
            fmt.Fprintf(&b, "  %s = insertvalue %s %s, %s %%%s, %d\n", cur, aggTy, prev, mapType(r.Values[i].Type), r.Values[i].ID, i)
            prev = cur
        }
        fmt.Fprintf(&b, "  ret %s %s\n", aggTy, prev)
        return b.String(), nil
    }
}
