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

func TestGenerateASM_MultiPath_Config_Normalized(t *testing.T) {
    m := ir.Module{Package: "p", Unit: "u.ami"}
    m.Pipelines = []ir.PipelineIR{{
        Name: "P",
        Steps: []ir.StepIR{{
            Node: "Collect",
            InMulti: &ir.MultiPathIR{
                Inputs: []edg.Spec{&edg.FIFO{MinCapacity: 1, MaxCapacity: 2, Backpressure: edg.BackpressureBlock, TypeName: "int"}},
                Merge: []ir.MergeOpIR{
                    {Name: "Sort", Raw: "ts,desc"},
                    {Name: "Stable", Raw: ""},
                    {Name: "Key", Raw: "id"},
                    {Name: "Dedup", Raw: "id"},
                    {Name: "Window", Raw: "5"},
                    {Name: "Watermark", Raw: "ts,10s"},
                    {Name: "Timeout", Raw: "1000"},
                    {Name: "Buffer", Raw: "3,dropOldest"},
                    {Name: "PartitionBy", Raw: "pid"},
                },
            },
        }},
    }}
    asm := GenerateASM(m)
    // Check mp_cfg contains normalized keys; order is sorted lexicographically
    want := []string{
        "mp_cfg ",
        "buffer.bp=dropOldest",
        "buffer.capacity=3",
        "dedup=true",
        "dedup.field=id",
        "key.field=id",
        "partitionBy.field=pid",
        "sort.field=ts",
        "sort.order=desc",
        "stable=true",
        "timeout.ms=1000",
        "watermark.field=ts",
        "watermark.lateness=10s",
        "window=5",
    }
    for _, s := range want {
        if !strings.Contains(asm, s) {
            t.Fatalf("missing %q in mp_cfg: %q", s, asm)
        }
    }
}

func TestGenerateASM_MultiPath_Config_ErrorPath_MultiPipelines(t *testing.T) {
    m := ir.Module{Package: "p", Unit: "u.ami"}
    m.Pipelines = []ir.PipelineIR{
        {
            Name: "P1",
            Steps: []ir.StepIR{{
                Node: "Collect",
                InMulti: &ir.MultiPathIR{
                    Inputs: []edg.Spec{&edg.FIFO{MinCapacity: 1, MaxCapacity: 1, Backpressure: edg.BackpressureBlock, TypeName: "int"}},
                    Merge:  []ir.MergeOpIR{{Name: "Buffer", Raw: "2,dropNewest"}},
                },
            }},
        },
        {
            Name: "P2",
            Steps: []ir.StepIR{{Node: "Ingress"}},
            ErrorSteps: []ir.StepIR{{
                Node: "Collect",
                InMulti: &ir.MultiPathIR{
                    Inputs: []edg.Spec{&edg.FIFO{MinCapacity: 0, MaxCapacity: 0, Backpressure: edg.BackpressureBlock, TypeName: "string"}},
                    Merge:  []ir.MergeOpIR{{Name: "Timeout", Raw: "500"}},
                },
            }},
        },
    }
    asm := GenerateASM(m)
    if !strings.Contains(asm, "pipeline P1") || !strings.Contains(asm, "pipeline P2") {
        t.Fatalf("missing pipeline headers: %q", asm)
    }
    // P1 normal path cfg
    if !strings.Contains(asm, "mp_cfg buffer.bp=dropNewest buffer.capacity=2") {
        t.Fatalf("expected P1 mp_cfg buffer settings in asm: %q", asm)
    }
    // P2 error path cfg
    if !strings.Contains(asm, ".error") { // label includes .error pipelines
        t.Fatalf("expected error path labels in asm: %q", asm)
    }
    if !strings.Contains(asm, "mp_cfg timeout.ms=500") {
        t.Fatalf("expected P2 error path timeout config in asm: %q", asm)
    }
}
