package exec

import (
    "context"
    "testing"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Ensures multiple transform stages run in order and filtering applies.
func TestExec_MultiTransform_Filtering(t *testing.T) {
    // Build module: ingress -> T1 -> T2 -> T3 -> egress
    edges := []edgeEntry{
        {Unit: "P", Pipeline: "P", From: "ingress", To: "T1"},
        {Unit: "P", Pipeline: "P", From: "T1", To: "T2"},
        {Unit: "P", Pipeline: "P", From: "T2", To: "T3"},
        {Unit: "P", Pipeline: "P", From: "T3", To: "egress"},
    }
    m := MakeModuleWithEdges(t, "app", "P", edges)
    // No collect needed for this test
    m.Pipelines[0].Collect = []ir.CollectSpec{}

    eng, err := NewEngineFromModule(m)
    if err != nil { t.Fatalf("engine: %v", err) }
    defer eng.Close()

    in := make(chan ev.Event, 16)
    // Produce 6 events with field i: 0..5
    go func(){
        for i := 0; i < 6; i++ { in <- ev.Event{Payload: map[string]any{"i": i}} }
        close(in)
    }()
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "drop_even", "add_field:flag", ExecOptions{})
    if err != nil { t.Fatalf("run: %v", err) }

    got := 0
    for e := range out {
        // Expect only odd i values due to filter
        if m, ok := e.Payload.(map[string]any); ok {
            if iv, ok := m["i"]; ok {
                var v int
                switch x := iv.(type) { case int: v = x; case float64: v = int(x); default: t.Fatalf("unexpected i type: %T", iv) }
                if v%2 == 0 { t.Fatalf("even value passed filter: %d", v) }
            } else { t.Fatalf("missing i in payload") }
        } else { t.Fatalf("unexpected payload type: %T", e.Payload) }
        got++
    }
    for range statsCh { /* ignore */ }
    if got != 3 { t.Fatalf("expected 3 odd outputs, got %d", got) }
}

