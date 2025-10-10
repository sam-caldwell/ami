package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "time"
    "context"

    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// helper to write a minimal edges.json with backpressure and capacity
func writeEdgesBP(t *testing.T, pkg, pipeline, from, to, bp string, cap int) {
    t.Helper()
    asm := filepath.Join("build", "debug", "asm", pkg)
    if err := os.MkdirAll(asm, 0o755); err != nil { t.Fatal(err) }
    idx := edgesIndex{Schema: "edges.v1", Package: pkg, Edges: []edgeEntry{{Pipeline: pipeline, From: from, To: to, Backpressure: bp, MaxCapacity: cap}}}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
}

func TestExec_Edges_Backpressure_DropNewest(t *testing.T) {
    // ingress->Transform default; Transform->egress dropNewest
    writeEdgesBP(t, "app", "P", "ingress", "Transform", "block", 0)
    // append second edge entry
    asm := filepath.Join("build", "debug", "asm", "app")
    b, _ := os.ReadFile(filepath.Join(asm, "edges.json"))
    var idx edgesIndex
    _ = json.Unmarshal(b, &idx)
    idx.Edges = append(idx.Edges, edgeEntry{Pipeline: "P", From: "Transform", To: "egress", Backpressure: "dropNewest", MaxCapacity: 2})
    bb, _ := json.Marshal(idx)
    _ = os.WriteFile(filepath.Join(asm, "edges.json"), bb, 0o644)

    // Build minimal IR module
    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P"}}}
    e, err := NewEngineFromModule(ir.Module{})
    if err != nil { t.Fatalf("engine: %v", err) }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Fast producer
    in := make(chan ev.Event, 16)
    go func(){ for i := 0; i < 10; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }; close(in) }()

    out, stats, err := e.RunPipelineWithStats(ctx, m, "P", in, nil, "", "", ExecOptions{})
    if err != nil { t.Fatalf("run: %v", err) }
    _ = stats
    // Delay consumption to force pressure on Transform->egress channel
    time.Sleep(50 * time.Millisecond)
    count := 0
    for range out { count++ }
    if count <= 0 || count > 2 { t.Fatalf("dropNewest expected <=2 items, got %d", count) }
}

func TestExec_Edges_Backpressure_DropOldest(t *testing.T) {
    writeEdgesBP(t, "app", "P", "ingress", "Transform", "block", 0)
    asm := filepath.Join("build", "debug", "asm", "app")
    b, _ := os.ReadFile(filepath.Join(asm, "edges.json"))
    var idx edgesIndex
    _ = json.Unmarshal(b, &idx)
    idx.Edges = append(idx.Edges, edgeEntry{Pipeline: "P", From: "Transform", To: "egress", Backpressure: "dropOldest", MaxCapacity: 2})
    bb, _ := json.Marshal(idx)
    _ = os.WriteFile(filepath.Join(asm, "edges.json"), bb, 0o644)

    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P"}}}
    e, err := NewEngineFromModule(ir.Module{})
    if err != nil { t.Fatalf("engine: %v", err) }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    in := make(chan ev.Event, 16)
    go func(){ for i := 0; i < 10; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }; close(in) }()
    out, _, err := e.RunPipelineWithStats(ctx, m, "P", in, nil, "", "", ExecOptions{})
    if err != nil { t.Fatalf("run: %v", err) }
    time.Sleep(50 * time.Millisecond)
    var got []int
    for e := range out {
        if m, ok := e.Payload.(map[string]any); ok {
            if v, ok := m["i"].(int); ok { got = append(got, v) }
        }
    }
    if len(got) != 2 { t.Fatalf("dropOldest expected 2 items, got %d (%v)", len(got), got) }
    if !(got[0] == 8 && got[1] == 9) { t.Fatalf("expected last two events [8 9], got %v", got) }
}

func TestExec_Edges_Backpressure_Block(t *testing.T) {
    writeEdgesBP(t, "app", "P", "ingress", "Transform", "block", 0)
    asm := filepath.Join("build", "debug", "asm", "app")
    b, _ := os.ReadFile(filepath.Join(asm, "edges.json"))
    var idx edgesIndex
    _ = json.Unmarshal(b, &idx)
    idx.Edges = append(idx.Edges, edgeEntry{Pipeline: "P", From: "Transform", To: "egress", Backpressure: "block", MaxCapacity: 2})
    bb, _ := json.Marshal(idx)
    _ = os.WriteFile(filepath.Join(asm, "edges.json"), bb, 0o644)

    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P"}}}
    e, err := NewEngineFromModule(ir.Module{})
    if err != nil { t.Fatalf("engine: %v", err) }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    in := make(chan ev.Event, 16)
    go func(){ for i := 0; i < 10; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }; close(in) }()
    out, _, err := e.RunPipelineWithStats(ctx, m, "P", in, nil, "", "", ExecOptions{})
    if err != nil { t.Fatalf("run: %v", err) }
    // delay start then consume all
    time.Sleep(50 * time.Millisecond)
    cnt := 0
    for range out { cnt++ }
    if cnt != 10 { t.Fatalf("block expected all items, got %d", cnt) }
}
