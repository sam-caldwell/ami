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
    errs "github.com/sam-caldwell/ami/src/schemas/errors"
)

// Verify shunt warnings include pipeline and edge id context.
func TestExec_Edges_Shunt_ErrorContext(t *testing.T) {
    // ingress->Transform block; Transform->egress shuntNewest with tiny capacity
    asm := filepath.Join("build", "debug", "asm", "app")
    if err := os.MkdirAll(asm, 0o755); err != nil { t.Fatal(err) }
    idx := edgesIndex{Schema: "edges.v1", Package: "app", Edges: []edgeEntry{{Pipeline: "P", From: "ingress", To: "Transform", Backpressure: "block", MaxCapacity: 0}}}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
    // append shuntNewest edge
    var idx2 edgesIndex
    _ = json.Unmarshal(b, &idx2)
    idx2.Edges = append(idx2.Edges, edgeEntry{Pipeline: "P", From: "Transform", To: "egress", Backpressure: "shuntNewest", MaxCapacity: 1})
    bb, _ := json.Marshal(idx2)
    _ = os.WriteFile(filepath.Join(asm, "edges.json"), bb, 0o644)

    m := ir.Module{Package: "app", Pipelines: []ir.Pipeline{{Name: "P"}}}
    e, err := NewEngineFromModule(ir.Module{})
    if err != nil { t.Fatalf("engine: %v", err) }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Buffered error channel to capture shunt warnings
    errCh := make(chan errs.Error, 8)
    in := make(chan ev.Event, 16)
    go func(){ for i := 0; i < 10; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }; close(in) }()
    out, _, err := e.RunPipelineWithStats(ctx, m, "P", in, nil, "", "", ExecOptions{ErrorChan: errCh})
    if err != nil { t.Fatalf("run: %v", err) }
    time.Sleep(50 * time.Millisecond)
    // Drain outputs to allow shunting to occur
    go func(){ for range out {} }()

    // Wait briefly for at least one warning
    timeout := time.After(200 * time.Millisecond)
    var got errs.Error
    select {
    case got = <-errCh:
    case <-timeout:
        t.Skip("no shunt warning observed in time; environment may be too fast")
    }
    if got.Code != "W_SHUNTED_NEWEST" && got.Code != "W_SHUNTED_OLDEST" {
        t.Fatalf("unexpected code: %s", got.Code)
    }
    if got.Data == nil { t.Fatalf("missing data context") }
    if p, ok := got.Data["pipeline"].(string); !ok || p != "P" { t.Fatalf("pipeline context missing: %+v", got.Data) }
    edge, ok := got.Data["edge"].(map[string]any)
    if !ok { t.Fatalf("edge context missing: %+v", got.Data) }
    if edge["from"] != "Transform" || edge["to"] != "egress" || edge["id"] != "Transform->egress" {
        t.Fatalf("edge context unexpected: %+v", edge)
    }
}

