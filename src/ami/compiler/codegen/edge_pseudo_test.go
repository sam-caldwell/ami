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
		Name: "P",
		Steps: []ir.StepIR{
			{
				Node: "Egress",
				In: &edg.FIFO{
					MinCapacity:  10,
					MaxCapacity:  20,
					Backpressure: edg.BackpressureBlock,
					TypeName:     "[]byte",
				},
			},
		},
	}}
	asm := GenerateASM(m)
	if !strings.Contains(asm, "pipeline P") {
		t.Fatalf("missing pipeline header: %q", asm)
	}
	const expectedAsmString = "edge_init label=P.step0.in kind=fifo min=10 max=20 bp=block type=[]byte"
	if !strings.Contains(asm, expectedAsmString) {
		t.Fatalf("missing edge_init pseudo: %q", asm)
	}
}

func TestGenerateASM_MultiPath_PseudoOps(t *testing.T) {
    m := ir.Module{Package: "p", Unit: "u.ami"}
    m.Pipelines = []ir.PipelineIR{{
        Name: "P",
        Steps: []ir.StepIR{{
            Node: "Collect",
            InMulti: &ir.MultiPathIR{
                Inputs: []edg.Spec{&edg.FIFO{MinCapacity: 1, MaxCapacity: 2, Backpressure: edg.BackpressureBlock, TypeName: "int"}},
                Merge:  []ir.MergeOpIR{{Name: "Sort", Raw: "event.ts,asc"}},
            },
        }},
    }}
    asm := GenerateASM(m)
    if !strings.Contains(asm, "mp_begin label=P.step0.in") { t.Fatalf("missing mp_begin: %q", asm) }
    if !strings.Contains(asm, "mp_input edge_init label=P.step0.in kind=fifo") { t.Fatalf("missing mp_input: %q", asm) }
    if !strings.Contains(asm, "mp_merge name=Sort args=event.ts,asc") { t.Fatalf("missing mp_merge: %q", asm) }
    if !strings.Contains(asm, "mp_end label=P.step0.in") { t.Fatalf("missing mp_end: %q", asm) }
}
