package llvm

import (
    "strings"
    "fmt"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func lowerPhi(p ir.Phi) string {
    ty := mapType(p.Result.Type)
    if ty == "" || ty == "void" { ty = "i64" }
    var parts []string
    for _, in := range p.Incomings {
        parts = append(parts, fmt.Sprintf("[ %%%s, %%%s ]", in.Value.ID, in.Label))
    }
    return fmt.Sprintf("  %%%s = phi %s %s\n", p.Result.ID, ty, strings.Join(parts, ", "))
}

