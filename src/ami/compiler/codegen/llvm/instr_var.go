package llvm

import (
    "fmt"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerVar emits a deterministic comment for variable declarations in the scaffold.
func lowerVar(v ir.Var) string {
    return fmt.Sprintf("  ; var %s : %s as %%%s\n", v.Name, mapType(v.Type), v.Result.ID)
}
