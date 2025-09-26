package codegen

import (
    "fmt"

    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
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

