package codegen

import (
    "fmt"
    "strings"

    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// edgeInitPseudo renders a single edge initialization pseudo-instruction for
// listings. Concrete implementations are specialized per payload type.
func edgeInitPseudo(pipe string, idx int, s edg.Spec) string {
    switch v := s.(type) {
    case *edg.FIFO:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=fifo min=%d max=%d bp=%s type=%s",
            pipe, idx, v.MinCapacity, v.MaxCapacity, v.Backpressure, v.TypeName)
    case *edg.LIFO:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=lifo min=%d max=%d bp=%s type=%s",
            pipe, idx, v.MinCapacity, v.MaxCapacity, v.Backpressure, v.TypeName)
    case *edg.Pipeline:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=pipeline upstream=%s min=%d max=%d bp=%s type=%s",
            pipe, idx, v.UpstreamName, v.MinCapacity, v.MaxCapacity, v.Backpressure, v.TypeName)
    default:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=%s", pipe, idx, s.Kind())
    }
}

// writeMultiPath emits no-op pseudo-ops for MultiPath scaffolding to aid
// future integration and debugging. It does not affect runtime semantics.
func writeMultiPath(b *strings.Builder, pipe string, idx int, st ir.StepIR) {
    b.WriteString("  mp_begin label=")
    b.WriteString(fmt.Sprintf("%s.step%d.in", pipe, idx))
    b.WriteString("\n")
    for _, in := range st.InMulti.Inputs {
        b.WriteString("  mp_input ")
        b.WriteString(edgeInitPseudo(pipe, idx, in))
        b.WriteString("\n")
    }
    for _, op := range st.InMulti.Merge {
        b.WriteString("  mp_merge name=")
        b.WriteString(op.Name)
        if op.Raw != "" { b.WriteString(" args="); b.WriteString(op.Raw) }
        b.WriteString("\n")
    }
    b.WriteString("  mp_end label=")
    b.WriteString(fmt.Sprintf("%s.step%d.in", pipe, idx))
    b.WriteString("\n")
}
