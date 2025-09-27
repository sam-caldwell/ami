package llvm

import (
    "fmt"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerAssign emits a deterministic comment for assignment in the scaffold.
func lowerAssign(a ir.Assign) string {
    return fmt.Sprintf("  ; assign %%%s = %%%s\n", a.DestID, a.Src.ID)
}

