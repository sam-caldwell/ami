package sem

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func canonicalDecoArgs(args []ast.Expr) string {
    if len(args) == 0 { return "" }
    // Preserve order; join with '|'
    out := make([]string, 0, len(args))
    for _, a := range args { out = append(out, decoArgText(a)) }
    return strings.Join(out, "|")
}

