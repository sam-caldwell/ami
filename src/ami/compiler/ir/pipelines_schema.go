package ir

import (
    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// ToPipelinesSchema converts lowered Pipelines into the public PipelinesV1 schema.
func (m Module) ToPipelinesSchema() sch.PipelinesV1 {
    out := sch.PipelinesV1{Schema: "pipelines.v1", Package: m.Package, File: m.Unit}
    for _, p := range m.Pipelines {
        sp := sch.PipelineV1{Name: p.Name}
        conv := func(steps []StepIR) []sch.PipelineStepV1 {
            var res []sch.PipelineStepV1
            for _, st := range steps {
                ps := sch.PipelineStepV1{Node: st.Node, Attrs: st.Attrs}
                // Attrs are not in StepIR yet; attach later if extended
                for _, w := range st.Workers {
                    ps.Workers = append(ps.Workers, sch.PipelineWorkerV1{
                        Name: w.Name, Kind: w.Kind, Origin: w.Origin, HasContext: w.HasContext, HasState: w.HasState,
                        Input: w.Input, OutputKind: w.OutputKind, Output: w.Output,
                    })
                }
                if st.In != nil {
                    ps.InEdge = toSchemaEdge(st.In)
                }
                res = append(res, ps)
            }
            return res
        }
        sp.Steps = conv(p.Steps)
        if len(p.ErrorSteps) > 0 {
            sp.ErrorSteps = conv(p.ErrorSteps)
        }
        out.Pipelines = append(out.Pipelines, sp)
    }
    return out
}

func toSchemaEdge(s edg.Spec) *sch.PipelineEdgeV1 {
    switch v := s.(type) {
    case *edg.FIFO:
        return &sch.PipelineEdgeV1{Kind: v.Kind(), MinCapacity: v.MinCapacity, MaxCapacity: v.MaxCapacity, Backpressure: string(v.Backpressure), Type: v.TypeName,
            Bounded: v.MaxCapacity > 0, Delivery: deliveryFromBP(string(v.Backpressure))}
    case *edg.LIFO:
        return &sch.PipelineEdgeV1{Kind: v.Kind(), MinCapacity: v.MinCapacity, MaxCapacity: v.MaxCapacity, Backpressure: string(v.Backpressure), Type: v.TypeName,
            Bounded: v.MaxCapacity > 0, Delivery: deliveryFromBP(string(v.Backpressure))}
    case *edg.Pipeline:
        return &sch.PipelineEdgeV1{Kind: v.Kind(), MinCapacity: v.MinCapacity, MaxCapacity: v.MaxCapacity, Backpressure: string(v.Backpressure), Type: v.TypeName, UpstreamName: v.UpstreamName,
            Bounded: v.MaxCapacity > 0, Delivery: deliveryFromBP(string(v.Backpressure))}
    default:
        return &sch.PipelineEdgeV1{Kind: s.Kind()}
    }
}

func deliveryFromBP(bp string) string {
    switch bp {
    case string(edg.BackpressureBlock):
        return "atLeastOnce"
    case string(edg.BackpressureDropOldest), string(edg.BackpressureDropNewest):
        return "bestEffort"
    default:
        return ""
    }
}
