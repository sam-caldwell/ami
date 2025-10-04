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

func TestWorker_ReturnsPayload_WrappedIntoEvent(t *testing.T) {
    // Arrange: write edges.json and pipelines.json for a simple ingress->Transform->egress path
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
    // pipelines: one Transform with worker=W
    type pipeList struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
    var pl pipeList
    pl.Pipelines = append(pl.Pipelines, struct{ Name string; Steps []struct{ Name string; Args []string } }{Name: "P", Steps: []struct{ Name string; Args []string }{{Name: "ingress"}, {Name: "Transform", Args: []string{"W"}}, {Name: "egress"}}})
    pb, _ := json.MarshalIndent(pl, "", "  ")
    if err := os.WriteFile(filepath.Join(dirIR, "u.pipelines.json"), pb, 0o644); err != nil { t.Fatal(err) }

    // Module with package only (edges/pipelines drive the path)
    m := ir.Module{Package: pkg, Pipelines: []ir.Pipeline{{Name: "P"}}}
    eng := &Engine{}
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    in := make(chan ev.Event, 4)
    // Send a couple of events with numeric payloads
    in <- ev.Event{Payload: 1}
    in <- ev.Event{Payload: 41}
    close(in)
    // Worker W: returns payload+1, nil error (bare payload form)
    opts := ExecOptions{Workers: map[string]func(ev.Event) (any, error){
        "W": func(e ev.Event) (any, error) {
            switch v := e.Payload.(type) {
            case int:
                return v + 1, nil
            default:
                return e.Payload, nil
            }
        },
    }}
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "", "", opts)
    if err != nil { t.Fatalf("run: %v", err) }
    // Collect outputs
    var got []int
    for e := range out {
        if n, ok := e.Payload.(int); ok { got = append(got, n) }
    }
    // Drain stats
    for range statsCh {}
    // Assert payloads were wrapped into events and incremented
    if len(got) != 2 || got[0] != 2 || got[1] != 42 {
        t.Fatalf("unexpected outputs: %v", got)
    }
    _ = time.Now()
}

