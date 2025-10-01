package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// WriteEdges writes build/debug/asm/<pkg>/edges.json with the provided entries.
func WriteEdges(t *testing.T, pkg, pipeline string, edges []edgeEntry) {
    t.Helper()
    dir := filepath.Join("build", "debug", "asm", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    idx := edgesIndex{Schema: "asm.v1", Package: pkg, Edges: edges}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(dir, "edges.json"), b, 0o644); err != nil { t.Fatalf("write edges: %v", err) }
}

// MakeModuleWithEdges constructs a minimal Module and writes edges for tests.
func MakeModuleWithEdges(t *testing.T, pkg, pipeline string, edges []edgeEntry) ir.Module {
    t.Helper()
    WriteEdges(t, pkg, pipeline, edges)
    return ir.Module{Package: pkg, Pipelines: []ir.Pipeline{{Name: pipeline}}}
}

// AttachCollect attaches a Collect spec with the given step name and plan to the
// specified pipeline in module m and returns the updated module.
func AttachCollect(m ir.Module, pipeline, step string, plan ir.MergePlan) ir.Module {
    for i := range m.Pipelines {
        if m.Pipelines[i].Name == pipeline {
            cp := plan // local copy for taking address safely
            m.Pipelines[i].Collect = append(m.Pipelines[i].Collect, ir.CollectSpec{Step: step, Merge: &cp})
        }
    }
    return m
}

// MakeCollectOnlyModule creates a module with a single Collect step and writes
// edges: ingress -> step -> egress.
func MakeCollectOnlyModule(t *testing.T, pkg, pipeline, step string, plan ir.MergePlan) ir.Module {
    t.Helper()
    edges := []edgeEntry{{Unit: pipeline, Pipeline: pipeline, From: "ingress", To: step}, {Unit: pipeline, Pipeline: pipeline, From: step, To: "egress"}}
    m := MakeModuleWithEdges(t, pkg, pipeline, edges)
    return AttachCollect(m, pipeline, step, plan)
}

// MakeTransformOnlyModule creates a module with transforms chained from ingress
// to egress and writes edges accordingly. No Collect steps are added.
func MakeTransformOnlyModule(t *testing.T, pkg, pipeline string, transforms []string) ir.Module {
    t.Helper()
    var edges []edgeEntry
    prev := "ingress"
    for _, name := range transforms {
        edges = append(edges, edgeEntry{Unit: pipeline, Pipeline: pipeline, From: prev, To: name})
        prev = name
    }
    edges = append(edges, edgeEntry{Unit: pipeline, Pipeline: pipeline, From: prev, To: "egress"})
    return MakeModuleWithEdges(t, pkg, pipeline, edges)
}

// MakeTransformAndCollectModule composes transforms followed by one Collect step.
func MakeTransformAndCollectModule(t *testing.T, pkg, pipeline string, transforms []string, collectStep string, plan ir.MergePlan) ir.Module {
    t.Helper()
    var edges []edgeEntry
    prev := "ingress"
    for _, name := range transforms {
        edges = append(edges, edgeEntry{Unit: pipeline, Pipeline: pipeline, From: prev, To: name})
        prev = name
    }
    edges = append(edges,
        edgeEntry{Unit: pipeline, Pipeline: pipeline, From: prev, To: collectStep},
        edgeEntry{Unit: pipeline, Pipeline: pipeline, From: collectStep, To: "egress"},
    )
    m := MakeModuleWithEdges(t, pkg, pipeline, edges)
    return AttachCollect(m, pipeline, collectStep, plan)
}

// MergePlanBuilder provides a fluent builder for common MergePlan variants.
type MergePlanBuilder struct{ p ir.MergePlan }

func NewMergePlan() MergePlanBuilder                       { return MergePlanBuilder{} }
func (b MergePlanBuilder) Stable(v bool) MergePlanBuilder  { b.p.Stable = v; return b }
func (b MergePlanBuilder) Key(field string) MergePlanBuilder { b.p.Key = field; return b }
func (b MergePlanBuilder) PartitionBy(field string) MergePlanBuilder { b.p.PartitionBy = field; return b }
func (b MergePlanBuilder) Buffer(capacity int, policy string) MergePlanBuilder {
    b.p.Buffer = ir.BufferPlan{Capacity: capacity, Policy: policy}; return b
}
func (b MergePlanBuilder) Window(n int) MergePlanBuilder    { b.p.Window = n; return b }
func (b MergePlanBuilder) TimeoutMs(ms int) MergePlanBuilder { b.p.TimeoutMs = ms; return b }
func (b MergePlanBuilder) SortAsc(field string) MergePlanBuilder {
    b.p.Sort = append(b.p.Sort, ir.SortKey{Field: field, Order: "asc"}); return b
}
func (b MergePlanBuilder) SortDesc(field string) MergePlanBuilder {
    b.p.Sort = append(b.p.Sort, ir.SortKey{Field: field, Order: "desc"}); return b
}
func (b MergePlanBuilder) Watermark(field string, latenessMs int) MergePlanBuilder {
    b.p.Watermark = &ir.Watermark{Field: field, LatenessMs: latenessMs}; return b
}
func (b MergePlanBuilder) Dedup(field string) MergePlanBuilder { b.p.DedupField = field; return b }
func (b MergePlanBuilder) LatePolicy(policy string) MergePlanBuilder { b.p.LatePolicy = policy; return b }
func (b MergePlanBuilder) Build() ir.MergePlan { return b.p }
