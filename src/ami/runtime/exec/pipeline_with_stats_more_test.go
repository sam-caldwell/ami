package exec

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestRunPipelineWithStats_FilterTransform_AndStats(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{Concurrency: 1, Schedule: "fifo"})
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    // create edges path
    asm := filepath.Join("build", "debug", "asm", "app")
    _ = os.MkdirAll(asm, 0o755)
    idx := edgesIndex{Schema: "asm.v1", Package: "app", Edges: []edgeEntry{
        {Pipeline: "P", From: "ingress", To: "Transform"},
        {Pipeline: "P", From: "Transform", To: "Collect"},
        {Pipeline: "P", From: "Collect", To: "egress"},
    }}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
    // module with collect
    mp := ir.MergePlan{Buffer: ir.BufferPlan{Capacity: 10, Policy: "block"}}
    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P", Collect: []ir.CollectSpec{{Step: "Collect", Merge: &mp}}}}}
    in := make(chan ev.Event, 8)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    // collect stage stats via channel from wrapper
    gotStats := make([]rmerge.Stats, 0, 4)
    emit := func(_ StageInfo, st rmerge.Stats){ gotStats = append(gotStats, st) }
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, emit, "drop_even", "add_field:flag", ExecOptions{})
    if err != nil { t.Fatalf("run pipelime stats: %v", err) }
    // send i=0..4 and then cancel; expect odd ones to pass through filter
    for i := 0; i < 5; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }
    time.Sleep(25*time.Millisecond)
    cancel()
    // drain
    outCnt := 0
    for range out { outCnt++ }
    // drain stats
    for range statsCh { /* consumed via emit too */ }
    if outCnt == 0 { t.Fatalf("expected some outputs after filter/transform") }
    if len(gotStats) == 0 { t.Fatalf("expected stage stats via emit") }
}

func TestRunPipelineWithStats_TimerSource(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{Concurrency: 1, Schedule: "fifo"})
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    // edges include Timer node
    asm := filepath.Join("build", "debug", "asm", "app")
    _ = os.MkdirAll(asm, 0o755)
    idx := edgesIndex{Schema: "asm.v1", Package: "app", Edges: []edgeEntry{
        {Pipeline: "P", From: "ingress", To: "Timer"},
        {Pipeline: "P", From: "Timer", To: "egress"},
    }}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
    m := ir.Module{Package: "app"}
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", nil, nil, "", "", ExecOptions{Sandbox: SandboxPolicy{AllowDevice: true}, TimerInterval: 5 * time.Millisecond, TimerCount: 2})
    if err != nil { t.Fatalf("run: %v", err) }
    // consume a few timer outputs then cancel
    got := 0
    for e := range out { _ = e; got++; if got >= 2 { break } }
    cancel()
    for range statsCh { /* drain */ }
    if got < 2 { t.Fatalf("expected >=2 timer outputs, got %d", got) }
}
