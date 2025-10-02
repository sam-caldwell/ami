package exec

import (
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
)

// Helper: convert IR MergePlan to runtime merge Plan.
func toRuntimePlan(p ir.MergePlan) rmerge.Plan {
    var rp rmerge.Plan
    rp.Stable = p.Stable
    for _, s := range p.Sort { rp.Sort = append(rp.Sort, rmerge.SortKey{Field: s.Field, Order: s.Order}) }
    rp.Key = p.Key
    rp.PartitionBy = p.PartitionBy
    rp.Buffer.Capacity = p.Buffer.Capacity
    rp.Buffer.Policy = p.Buffer.Policy
    rp.Window = p.Window
    rp.TimeoutMs = p.TimeoutMs
    rp.LatePolicy = p.LatePolicy
    if p.DedupField != "" { rp.Dedup.Field = p.DedupField }
    if p.Watermark != nil { rp.Watermark = &rmerge.Watermark{Field: p.Watermark.Field, LatenessMs: p.Watermark.LatenessMs} }
    return rp
}

