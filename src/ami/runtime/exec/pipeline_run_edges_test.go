package exec

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestRunPipeline_EdgesPath(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{Concurrency: 1, Schedule: "fifo"})
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    // Write edges.json under ./build/debug/asm/app
    asm := filepath.Join("build", "debug", "asm", "app")
    _ = os.MkdirAll(asm, 0o755)
    idx := edgesIndex{Schema: "asm.v1", Package: "app", Edges: []edgeEntry{
        {Pipeline: "P", From: "ingress", To: "Transform"},
        {Pipeline: "P", From: "Transform", To: "Collect"},
        {Pipeline: "P", From: "Collect", To: "egress"},
    }}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
    m := ir.Module{Package: "app"}
    mp := ir.MergePlan{Buffer: ir.BufferPlan{Capacity: 5, Policy: "block"}}
    m.Pipelines = []ir.Pipeline{{Name: "P", Collect: []ir.CollectSpec{{Step: "Collect", Merge: &mp}}}}
    in := make(chan ev.Event, 8)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    out, err := eng.RunPipeline(ctx, m, "P", in)
    if err != nil { t.Fatalf("run pipeline: %v", err) }
    for i := 0; i < 2; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }
    time.Sleep(10*time.Millisecond)
    cancel()
    cnt := 0
    for range out { cnt++ }
    if cnt == 0 { t.Fatalf("expected some output via edges path") }
}

