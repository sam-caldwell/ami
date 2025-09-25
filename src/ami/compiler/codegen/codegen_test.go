package codegen

import (
	edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
	"github.com/sam-caldwell/ami/src/ami/compiler/ir"
	"strings"
	"testing"
)

func TestGenerateASM_ContainsFunctionLabels(t *testing.T) {
	m := ir.Module{Package: "p", Unit: "u.ami", Functions: []ir.Function{{Name: "main"}, {Name: "helper"}}}
	asm := GenerateASM(m)
	if !strings.Contains(asm, "fn_main:") || !strings.Contains(asm, "fn_helper:") {
		t.Fatalf("labels missing: %q", asm)
	}
}

func TestGenerateASM_Golden(t *testing.T) {
	m := ir.Module{Package: "p", Unit: "u.ami", Functions: []ir.Function{{Name: "a"}, {Name: "b"}}}
	got := GenerateASM(m)
	want := "; AMI-IR assembly\n; package p unit u.ami\nfn_a:\n  ret\nfn_b:\n  ret\n"
	if got != want {
		t.Fatalf("\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestGenerateASM_EmitsEdgeInitPseudo(t *testing.T) {
	m := ir.Module{Package: "p", Unit: "u.ami"}
	// Add a pipeline with one step that has a FIFO edge
	m.Pipelines = []ir.PipelineIR{{
		Name:  "P",
		Steps: []ir.StepIR{{Node: "Egress", In: &edg.FIFO{MinCapacity: 10, MaxCapacity: 20, Backpressure: edg.BackpressureBlock, TypeName: "[]byte"}}},
	}}
	asm := GenerateASM(m)
	if !strings.Contains(asm, "pipeline P") {
		t.Fatalf("missing pipeline header: %q", asm)
	}
	if !strings.Contains(asm, "edge_init label=P.step0.in kind=fifo min=10 max=20 bp=block type=[]byte") {
		t.Fatalf("missing edge_init pseudo: %q", asm)
	}
}
