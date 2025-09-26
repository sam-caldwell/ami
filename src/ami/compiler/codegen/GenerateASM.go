package codegen

import (
    "fmt"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// GenerateASM renders a human-readable assembly-like listing from the IR module.
// It includes package/unit headers, function labels, and edge init pseudo-ops
// for pipelines to make initialization concrete in listings.
func GenerateASM(m ir.Module) string {
    var b strings.Builder
    b.WriteString("; AMI-IR assembly\n")
    b.WriteString("; package ")
    b.WriteString(m.Package)
    b.WriteString(" unit ")
    b.WriteString(m.Unit)
    b.WriteString("\n")
    if m.Concurrency > 0 {
        b.WriteString(fmt.Sprintf("; concurrency %d\n", m.Concurrency))
    }
    if m.Backpressure != "" {
        b.WriteString("; backpressure ")
        b.WriteString(m.Backpressure)
        b.WriteString("\n")
    }
    if len(m.Capabilities) > 0 {
        b.WriteString("; capabilities ")
        b.WriteString(strings.Join(m.Capabilities, ","))
        b.WriteString("\n")
    }
    if m.Trust != "" {
        b.WriteString("; trust ")
        b.WriteString(m.Trust)
        b.WriteString("\n")
    }
    for _, fn := range m.Functions {
        b.WriteString("fn_")
        b.WriteString(fn.Name)
        b.WriteString(":\n  ret\n")
    }
    for _, p := range m.Pipelines {
        b.WriteString("; pipeline ")
        b.WriteString(p.Name)
        b.WriteString("\n")
        for i, st := range p.Steps {
            if st.In != nil {
                b.WriteString("  ")
                b.WriteString(edgeInitPseudo(p.Name, i, st.In))
                b.WriteString("\n")
                continue
            }
            if st.InMulti != nil {
                writeMultiPath(&b, p.Name, i, st)
            }
        }
        for i, st := range p.ErrorSteps {
            if st.In != nil {
                b.WriteString("  ")
                b.WriteString(edgeInitPseudo(p.Name+".error", i, st.In))
                b.WriteString("\n")
                continue
            }
            if st.InMulti != nil {
                writeMultiPath(&b, p.Name+".error", i, st)
            }
        }
    }
    return b.String()
}
