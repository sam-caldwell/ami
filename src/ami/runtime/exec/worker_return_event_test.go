package exec

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Verify a worker that returns an Event directly is forwarded unchanged.
func TestWorker_ReturnsEvent_Forwarded(t *testing.T) {
    pkg := "app"
    dirAsm := filepath.Join("build", "debug", "asm", pkg)
    dirIR := filepath.Join("build", "debug", "ir", pkg)
    _ = os.MkdirAll(dirAsm, 0o755)
    _ = os.MkdirAll(dirIR, 0o755)
    // edges: ingress -> Transform -> egress
    type edgeEntry struct{ Schema, Package string; Edges []struct{ Pipeline, From, To string } }
    eidx := edgeEntry{Schema: "asm.v1", Package: pkg}
    eidx.Edges = append(eidx.Edges, struct{ Pipeline, From, To string }{"P", "ingress", "Transform"})
    eidx.Edges = append(eidx.Edges, struct{ Pipeline, From, To string }{"P", "Transform", "egress"})
    b, _ := json.MarshalIndent(eidx, "", "  ")
    if err := os.WriteFile(filepath.Join(dirAsm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
    // pipelines: Transform with worker=W
    type pipeList struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
    var pl pipeList
    pl.Pipelines = append(pl.Pipelines, struct{ Name string; Steps []struct{ Name string; Args []string } }{Name: "P", Steps: []struct{ Name string; Args []string }{{Name: "ingress"}, {Name: "Transform", Args: []string{"W"}}, {Name: "egress"}}})
    pb, _ := json.MarshalIndent(pl, "", "  ")
    if err := os.WriteFile(filepath.Join(dirIR, "u.pipelines.json"), pb, 0o644); err != nil { t.Fatal(err) }

    m := ir.Module{Package: pkg, Pipelines: []ir.Pipeline{{Name: "P"}}}
    eng := &Engine{}
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    in := make(chan ev.Event, 1)
    in <- ev.Event{Payload: map[string]any{"i": 10}}
    close(in)
    // Worker returns an Event with modified payload
    opts := ExecOptions{Workers: map[string]func(ev.Event) (any, error){
        "W": func(e ev.Event) (any, error) {
            e.Payload = map[string]any{"i": 99}
            return e, nil
        },
    }}
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "", "", opts)
    if err != nil { t.Fatalf("run: %v", err) }
    var got []int
    for e := range out {
        m, _ := e.Payload.(map[string]any)
        switch x := m["i"].(type) { case int: got = append(got, x); case float64: got = append(got, int(x)) }
    }
    for range statsCh {}
    if len(got) != 1 || got[0] != 99 {
        t.Fatalf("unexpected outputs: %v", got)
    }
}

