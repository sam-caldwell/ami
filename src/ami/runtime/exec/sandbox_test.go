package exec

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "testing"
    "time"
)

func TestSandbox_TimerDenied(t *testing.T) {
    // Minimal module with package/pipeline metadata
    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P"}}}
    // Write edges to include a Timer node so executor uses IR-defined trigger
    mustWriteEdges(t, "app", "P", []edgeEntry{
        {Unit: "P", Pipeline: "P", From: "ingress", To: "Timer"},
        {Unit: "P", Pipeline: "P", From: "Timer", To: "egress"},
    })
    eng, err := NewEngineFromModule(m)
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    in := make(chan ev.Event)
    close(in) // unused
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()
    // Deny device => timer should be blocked
    _, _, err = eng.RunPipelineWithStats(ctx, m, "P", in, nil, "none", "none", ExecOptions{TimerInterval: 5 * time.Millisecond, TimerCount: 1, Sandbox: SandboxPolicy{AllowDevice: false, AllowFS: true, AllowNet: true}})
    if err == nil { t.Fatalf("expected sandbox denial error") }
    if _, ok := err.(ErrSandboxDenied); !ok { t.Fatalf("expected ErrSandboxDenied; got %T %v", err, err) }
}

func TestSandbox_TimerAllowed(t *testing.T) {
    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P"}}}
    mustWriteEdges(t, "app", "P", []edgeEntry{
        {Unit: "P", Pipeline: "P", From: "ingress", To: "Timer"},
        {Unit: "P", Pipeline: "P", From: "Timer", To: "egress"},
    })
    eng, err := NewEngineFromModule(m)
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    in := make(chan ev.Event)
    close(in)
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "none", "none", ExecOptions{TimerInterval: 5 * time.Millisecond, TimerCount: 3, Sandbox: SandboxPolicy{AllowDevice: true}})
    if err != nil { t.Fatalf("run: %v", err) }
    // Drain outputs and ensure at least one event was produced
    count := 0
    for e := range out { _ = e; count++ }
    // Drain stats
    for range statsCh { /* ignore */ }
    if count == 0 { t.Fatalf("expected timer to emit events; got 0") }
}

// mustWriteEdges writes a minimal build/debug/asm/<pkg>/edges.json with provided edges.
func mustWriteEdges(t *testing.T, pkg, pipeline string, edges []edgeEntry) {
    t.Helper()
    dir := filepath.Join("build", "debug", "asm", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    idx := edgesIndex{Schema: "asm.v1", Package: pkg, Edges: edges}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(dir, "edges.json"), b, 0o644); err != nil { t.Fatalf("write edges: %v", err) }
}
