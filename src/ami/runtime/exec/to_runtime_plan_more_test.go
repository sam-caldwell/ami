package exec

import (
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func Test_toRuntimePlan_ConvertsAllFields(t *testing.T) {
    p := ir.MergePlan{Stable: true, Sort: []ir.SortKey{{Field: "ts", Order: "asc"}}, Key: "k", PartitionBy: "p", Buffer: ir.BufferPlan{Capacity: 3, Policy: "block"}, Window: 2, TimeoutMs: 5, LatePolicy: "accept", DedupField: "id", Watermark: &ir.Watermark{Field: "ts", LatenessMs: 10}}
    rp := toRuntimePlan(p)
    if !rp.Stable || len(rp.Sort) != 1 || rp.Key != "k" || rp.PartitionBy != "p" { t.Fatalf("basic fields not set: %+v", rp) }
    if rp.Buffer.Capacity != 3 || rp.Buffer.Policy != "block" || rp.Window != 2 || rp.TimeoutMs != 5 { t.Fatalf("buffer/window/timeout not set: %+v", rp) }
    if rp.Dedup.Field != "id" { t.Fatalf("dedup not set: %+v", rp) }
    if rp.Watermark == nil || rp.Watermark.Field != "ts" || rp.Watermark.LatenessMs != 10 { t.Fatalf("watermark not set: %+v", rp) }
}

