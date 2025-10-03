package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestEmitModuleLLVM_EmbedsModuleMeta(t *testing.T) {
    m := ir.Module{Package: "app", Concurrency: 4, Backpressure: "block", Schedule: "fair", TelemetryEnabled: true, Capabilities: []string{"io", "net"}, TrustLevel: "trusted"}
    m.Pipelines = []ir.Pipeline{
        {
            Name: "P",
            Collect: []ir.CollectSpec{
                {Step: "Collect", Merge: &ir.MergePlan{Buffer: ir.BufferPlan{Capacity: 10, Policy: "dropNewest"}, Sort: []ir.SortKey{{Field: "ts", Order: "asc"}}, Stable: true, Window: 5, Watermark: &ir.Watermark{Field: "ts"}, TimeoutMs: 100, PartitionBy: "k", Key: "k"}},
            },
        },
    }
    m.ErrorPipes = []ir.ErrorPipeline{{Pipeline: "P", Steps: []string{"ingress","Transform","egress"}}}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "@ami_meta_json = private constant") { t.Fatalf("missing meta global: %s", out) }
    // Spot-check a few key fields in the embedded JSON (escaped form in c-string)
    if !strings.Contains(out, "\\22schema\\22:\\22ami.meta.v1\\22") { t.Fatalf("schema missing: %s", out) }
    if !strings.Contains(out, "\\22package\\22:\\22app\\22") { t.Fatalf("package missing: %s", out) }
    if !strings.Contains(out, "\\22concurrency\\22:4") { t.Fatalf("concurrency missing: %s", out) }
    if !strings.Contains(out, "\\22backpressure\\22:\\22block\\22") { t.Fatalf("backpressure missing: %s", out) }
    if !strings.Contains(out, "\\22capabilities\\22:[\\22io\\22,\\22net\\22]") { t.Fatalf("capabilities missing: %s", out) }
    if !strings.Contains(out, "\\22pipelines\\22:") { t.Fatalf("pipelines missing: %s", out) }
    if !strings.Contains(out, "\\22errorPipelines\\22:") { t.Fatalf("errorPipelines missing: %s", out) }
}

