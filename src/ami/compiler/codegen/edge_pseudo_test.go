package codegen

import (
    "strings"
    "testing"

    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestGenerateASM_EmitsEdgeInitPseudo(t *testing.T) {
    m := ir.Module{Package: "p", Unit: "u.ami"}
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

