package exec

import (
    "context"
    "testing"
    "time"
    "github.com/sam-caldwell/ami/src/testutil"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Verifies an IR-triggered Timer flows through a transform and Collect, emitting outputs.
func TestExec_Timer_Transform_Collect(t *testing.T) {
    // Build module: ingress -> Timer -> Xform -> Collect -> egress
    edges := []edgeEntry{
        {Unit: "P", Pipeline: "P", From: "ingress", To: "Timer"},
        {Unit: "P", Pipeline: "P", From: "Timer", To: "Xform"},
        {Unit: "P", Pipeline: "P", From: "Xform", To: "Collect"},
        {Unit: "P", Pipeline: "P", From: "Collect", To: "egress"},
    }
    m := MakeModuleWithEdges(t, "app", "P", edges)
    // Attach a simple MergePlan: sort by ts ascending, small window to flush promptly
    m.Pipelines[0].Collect = []ir.CollectSpec{{Step: "Collect", Merge: &ir.MergePlan{Sort: []ir.SortKey{{Field: "ts", Order: "asc"}}, Window: 1}}}

    eng, err := NewEngineFromModule(m)
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()

    in := make(chan ev.Event)
    close(in) // unused (timer provides source)
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "none", "add_field:flag", ExecOptions{TimerInterval: 5 * time.Millisecond, TimerCount: testutil.ScaleInt(3), Sandbox: SandboxPolicy{AllowDevice: true}})
    if err != nil { t.Fatalf("run: %v", err) }

    count := 0
    for e := range out {
        // Expect transform added flag=true
        if m, ok := e.Payload.(map[string]any); ok {
            if v, ok := m["flag"].(bool); !ok || !v { t.Fatalf("missing transform flag in payload: %+v", e.Payload) }
        } else { t.Fatalf("unexpected payload type: %T", e.Payload) }
        count++
    }
    // Drain stats (not asserted here)
    for range statsCh { /* no-op */ }
    if count != testutil.ScaleInt(3) { t.Fatalf("expected %d outputs from timer, got %d", testutil.ScaleInt(3), count) }
}
