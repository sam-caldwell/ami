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
                } else if st.InMulti != nil {
                    // Convert MultiPath IR into schema
                    var inputs []sch.PipelineEdgeV1
                    for _, e := range st.InMulti.Inputs {
                        if pe := toSchemaEdge(e); pe != nil { inputs = append(inputs, *pe) }
                    }
                    var merge []sch.MergeOpV1
                    for _, op := range st.InMulti.Merge {
                        merge = append(merge, sch.MergeOpV1{Name: op.Name, Raw: op.Raw})
                    }
                    mp := &sch.MultiPathV1{Inputs: inputs, Merge: merge}
                    if st.InMulti.Config != nil {
                        c := st.InMulti.Config
                        mp.MergeConfig = &sch.MergeConfigV1{
                            SortField: c.SortField, SortOrder: c.SortOrder, Stable: c.Stable,
                            Key: c.Key, Dedup: c.Dedup, DedupField: c.DedupField, Window: c.Window,
                            WatermarkField: c.WatermarkField, WatermarkLateness: c.WatermarkLateness,
                            TimeoutMs: c.TimeoutMs, BufferCapacity: c.BufferCapacity, BufferBackpressure: c.BufferBackpressure,
                            PartitionBy: c.PartitionBy,
                        }
                    }
                    ps.InEdge = &sch.PipelineEdgeV1{Kind: "edge.MultiPath", MultiPath: mp}
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
