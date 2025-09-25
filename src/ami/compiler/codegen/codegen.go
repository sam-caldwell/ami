package codegen

import (
	"fmt"
	edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
	"github.com/sam-caldwell/ami/src/ami/compiler/ir"
	"strings"
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
	// pragma-derived attributes (subset)
	if m.Concurrency > 0 {
		b.WriteString(fmt.Sprintf("; concurrency %d\n", m.Concurrency))
	}
	if m.Backpressure != "" {
		b.WriteString("; backpressure ")
		b.WriteString(m.Backpressure)
		b.WriteString("\n")
	}
	// scheduling and telemetry pragmas removed from language; no emission
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
	// Emit declarative pipeline/edge initialization pseudo-ops (skeleton)
	for _, p := range m.Pipelines {
		b.WriteString("; pipeline ")
		b.WriteString(p.Name)
		b.WriteString("\n")
		for i, st := range p.Steps {
			if st.In == nil {
				continue
			}
			b.WriteString("  ")
			b.WriteString(edgeInitPseudo(p.Name, i, st.In))
			b.WriteString("\n")
		}
		for i, st := range p.ErrorSteps {
			if st.In == nil {
				continue
			}
			b.WriteString("  ")
			b.WriteString(edgeInitPseudo(p.Name+".error", i, st.In))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// edgeInitPseudo renders a single edge initialization pseudo-instruction.
// This is a codegen skeleton to make the high-performance path concrete in
// assembly listings. Concrete implementations are specialized per payload type.
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
