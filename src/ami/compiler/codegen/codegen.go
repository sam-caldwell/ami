package codegen

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// GenerateASM renders a human-readable assembly-like listing from the IR module.
func GenerateASM(m ir.Module) string {
    var b strings.Builder
    b.WriteString("; AMI-IR assembly\n")
    b.WriteString("; package ")
    b.WriteString(m.Package)
    b.WriteString(" unit ")
    b.WriteString(m.Unit)
    b.WriteString("\n")
    for _, fn := range m.Functions {
        b.WriteString("fn_")
        b.WriteString(fn.Name)
        b.WriteString(":\n  ret\n")
    }
    return b.String()
}

