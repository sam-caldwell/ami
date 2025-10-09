package exec

import (
    "context"
    "testing"
    "time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Validate that a pipeline with a GpuDispatch transform integrates with scheduling,
// honors sandbox policy, and forwards events deterministically.
func TestExec_GpuDispatch_Transform_PassThrough_WithStats(t *testing.T) {
    // Build module: ingress -> GpuDispatch -> egress
    edges := []edgeEntry{
        {Unit: "P", Pipeline: "P", From: "ingress", To: "GpuDispatch"},
        {Unit: "P", Pipeline: "P", From: "GpuDispatch", To: "egress"},
    }
    m := MakeModuleWithEdges(t, "app", "P", edges)
    eng, err := NewEngineFromModule(m)
    if err != nil { t.Fatalf("engine: %v", err) }
    ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
    defer cancel()
    // Prepare input
    in := make(chan ev.Event, 3)
    in <- ev.Event{Payload: 1}; in <- ev.Event{Payload: 2}; in <- ev.Event{Payload: 3}
    close(in)
    // Run with device allowed in sandbox
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "none", "none", ExecOptions{Sandbox: SandboxPolicy{AllowDevice: true}})
    if err != nil { t.Fatalf("run: %v", err) }
    var got []int
    for e := range out {
        switch x := e.Payload.(type) { case int: got = append(got, x); case float64: got = append(got, int(x)) }
    }
    // Drain stats
    for range statsCh {}
    if len(got) != 3 || got[0] != 1 || got[2] != 3 { t.Fatalf("unexpected outputs: %v", got) }
}

func TestExec_GpuDispatch_SandboxDenied(t *testing.T) {
    // Build module: ingress -> GpuDispatch -> egress
    edges := []edgeEntry{
        {Unit: "P", Pipeline: "P", From: "ingress", To: "GpuDispatch"},
        {Unit: "P", Pipeline: "P", From: "GpuDispatch", To: "egress"},
    }
    m := MakeModuleWithEdges(t, "app", "P", edges)
    eng, _ := NewEngineFromModule(m)
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    in := make(chan ev.Event, 1)
    close(in)
    _, _, err := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "none", "none", ExecOptions{Sandbox: SandboxPolicy{AllowDevice: false}})
    if err == nil { t.Fatalf("expected sandbox denied error") }
}

