package llvm

import (
    "fmt"
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
        return "", fmt.Errorf("multi-value return not supported in LLVM emitter scaffold")
    }
}

