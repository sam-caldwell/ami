package exec

import (
    "context"
    "testing"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func TestRunPipelineWithStats_FallbackPath(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{Concurrency: 1, Schedule: "fifo"})
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()
    mp := ir.MergePlan{}
    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P", Collect: []ir.CollectSpec{{Step: "Collect", Merge: &mp}}}}}
    in := make(chan ev.Event, 4)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    got := make([]rmerge.Stats, 0, 2)
    emit := func(_ StageInfo, st rmerge.Stats){ got = append(got, st) }
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, emit, "", "", ExecOptions{})
    if err != nil { t.Fatalf("run: %v", err) }
    in <- ev.Event{Payload: map[string]any{"x": 1}}
    time.Sleep(10*time.Millisecond)
    cancel()
    for range out { /* drain */ }
    for range statsCh { /* drain */ }
    if len(got) == 0 { t.Fatalf("expected stage stats emit") }
}

func TestRunPipelineWithStats_NoNodes_EmitsEgressStats(t *testing.T) {
    eng, _ := NewEngineFromModule(ir.Module{Concurrency: 1})
    defer eng.Close()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    out, statsCh, err := eng.RunPipelineWithStats(ctx, ir.Module{Package: "app"}, "Unknown", nil, nil, "", "", ExecOptions{})
    if err != nil { t.Fatalf("err: %v", err) }
    cancel()
    <-statsCh // should receive egress stats then close
    // out should be the input (nil) channel; ensure not nil to exercise path
    _ = out
}
