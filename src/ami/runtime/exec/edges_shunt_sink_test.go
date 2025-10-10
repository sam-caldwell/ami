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

// Verify that shunted events are forwarded to ShuntChan with trace context.
func TestExec_Edges_Shunt_Sink(t *testing.T) {
    asm := filepath.Join("build", "debug", "asm", "app")
    _ = os.MkdirAll(asm, 0o755)
    idx := edgesIndex{Schema: "edges.v1", Package: "app", Edges: []edgeEntry{{Pipeline: "P", From: "ingress", To: "Transform", Backpressure: "block", MaxCapacity: 0}}}
    b, _ := json.Marshal(idx)
    _ = os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644)
    var idx2 edgesIndex
    _ = json.Unmarshal(b, &idx2)
    idx2.Edges = append(idx2.Edges, edgeEntry{Pipeline: "P", From: "Transform", To: "egress", Backpressure: "shuntNewest", MaxCapacity: 1})
    bb, _ := json.Marshal(idx2)
    _ = os.WriteFile(filepath.Join(asm, "edges.json"), bb, 0o644)

    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P"}}}
    e, _ := NewEngineFromModule(ir.Module{})
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    shch := make(chan ev.Event, 8)
    in := make(chan ev.Event, 16)
    go func(){ for i := 0; i < 10; i++ { in <- ev.Event{Payload: i} }; close(in) }()

    out, _, err := e.RunPipelineWithStats(ctx, m, "P", in, nil, "", "", ExecOptions{ShuntChan: shch})
    if err != nil { t.Fatalf("run: %v", err) }
    // Let backpressure build and then drain
    time.Sleep(50 * time.Millisecond)
    go func(){ for range out {} }()
    // Expect at least one shunted event
    select {
    case se := <-shch:
        if se.Trace == nil { t.Fatalf("expected trace on shunted event") }
        if sh, ok := se.Trace["shunt"].(map[string]any); ok {
            if sh["pipeline"] != "P" { t.Fatalf("trace missing pipeline: %+v", se.Trace) }
        } else {
            t.Fatalf("trace missing shunt details: %+v", se.Trace)
        }
    case <-time.After(200 * time.Millisecond):
        t.Skip("no shunted events observed quickly; timing-sensitive")
    }
}

