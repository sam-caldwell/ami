package root_test

import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// Ensure pipelines.v1 includes step Attrs for typical cases
// like minWorkers and in=edge.* while also preserving structured InEdge.
func TestBuild_PipelinesIR_IncludesStepAttrs(t *testing.T) {
    t.Setenv("AMI_SEM_DIAGS", "0")
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // minimal workspace
    ws := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env: []
  linker: {}
  linter: {}
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
        t.Fatal(err)
    }
    if err := os.MkdirAll("src", 0o755); err != nil {
        t.Fatal(err)
    }

    // Worker + pipelines with attrs: worker, min/max workers, onError, capabilities
    // and a variety of edge specs (FIFO/LIFO/Pipeline)
    src := `package main
func doThing(ev Event<string>) (Event<string>, error) { return ev, nil }
pipeline P {
  Ingress(cfg).
  Transform(worker=doThing,minWorkers=2,maxWorkers=4,onError=drop,capabilities=net).
  Egress(in=edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=string))
}

pipeline Q {
  Ingress(cfg).
  Egress(in=edge.LIFO(minCapacity=0,maxCapacity=3,backpressure=dropNewest,type=string))
}

pipeline R {
  Ingress(cfg).
  Egress(in=edge.Pipeline(name="P",minCapacity=5,maxCapacity=5,backpressure=dropOldest,type=string))
}`
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
        t.Fatal(err)
    }

    // Run build --verbose
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "build", "--verbose"}
    _ = captureStdoutBuild(t, func() { _ = rootcmd.Execute() })

    // Read pipelines JSON
    path := filepath.Join("build", "debug", "ir", "main", "main.ami.pipelines.json")
    b, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("missing pipelines json: %v", err)
    }
    var p sch.PipelinesV1
    if err := json.Unmarshal(b, &p); err != nil {
        t.Fatalf("unmarshal pipelines: %v", err)
    }
    if err := p.Validate(); err != nil {
        t.Fatalf("pipelines schema validate: %v", err)
    }
    // Find pipeline P
    var pipe *sch.PipelineV1
    for i := range p.Pipelines {
        if p.Pipelines[i].Name == "P" {
            pipe = &p.Pipelines[i]
            break
        }
    }
    if pipe == nil || len(pipe.Steps) < 3 {
        t.Fatalf("expected pipeline P with 3 steps; got %+v", pipe)
    }

    // Transform step should carry worker/min/max/onError/capabilities in Attrs
    tr := pipe.Steps[1]
    if tr.Attrs == nil || tr.Attrs["worker"] != "doThing" || tr.Attrs["minWorkers"] != "2" || tr.Attrs["maxWorkers"] != "4" || tr.Attrs["onError"] != "drop" || tr.Attrs["capabilities"] != "net" {
        t.Fatalf("expected step attrs worker/doThing,minWorkers=2,maxWorkers=4,onError=drop,capabilities=net; got %+v", tr.Attrs)
    }

    // Egress step should carry raw in=edge.FIFO(...) in Attrs and structured InEdge
    eg := pipe.Steps[2]
    if eg.Attrs == nil || !strings.HasPrefix(strings.TrimSpace(eg.Attrs["in"]), "edge.FIFO(") {
        t.Fatalf("expected egress attr in=edge.FIFO(...); got %+v", eg.Attrs)
    }
    if eg.InEdge == nil || eg.InEdge.Kind != "edge.FIFO" || eg.InEdge.MinCapacity != 1 || eg.InEdge.MaxCapacity != 2 || eg.InEdge.Backpressure != "block" || eg.InEdge.Type != "string" {
        t.Fatalf("unexpected structured inEdge: %+v", eg.InEdge)
    }

    // Validate LIFO in pipeline Q
    var q *sch.PipelineV1
    for i := range p.Pipelines {
        if p.Pipelines[i].Name == "Q" { q = &p.Pipelines[i]; break }
    }
    if q == nil || len(q.Steps) < 2 {
        t.Fatalf("expected pipeline Q with at least 2 steps; got %+v", q)
    }
    qeg := q.Steps[len(q.Steps)-1]
    if qeg.Attrs == nil || !strings.HasPrefix(strings.TrimSpace(qeg.Attrs["in"]), "edge.LIFO(") {
        t.Fatalf("expected Q egress attr in=edge.LIFO(...); got %+v", qeg.Attrs)
    }
    if qeg.InEdge == nil || qeg.InEdge.Kind != "edge.LIFO" || qeg.InEdge.MinCapacity != 0 || qeg.InEdge.MaxCapacity != 3 || qeg.InEdge.Backpressure != "dropNewest" || qeg.InEdge.Type != "string" {
        t.Fatalf("unexpected Q structured inEdge: %+v", qeg.InEdge)
    }

    // Validate Pipeline edge in pipeline R
    var r *sch.PipelineV1
    for i := range p.Pipelines {
        if p.Pipelines[i].Name == "R" { r = &p.Pipelines[i]; break }
    }
    if r == nil || len(r.Steps) < 2 {
        t.Fatalf("expected pipeline R with at least 2 steps; got %+v", r)
    }
    reg := r.Steps[len(r.Steps)-1]
    if reg.Attrs == nil || !strings.HasPrefix(strings.TrimSpace(reg.Attrs["in"]), "edge.Pipeline(") {
        t.Fatalf("expected R egress attr in=edge.Pipeline(...); got %+v", reg.Attrs)
    }
    if reg.InEdge == nil || reg.InEdge.Kind != "edge.Pipeline" || reg.InEdge.MinCapacity != 5 || reg.InEdge.MaxCapacity != 5 || reg.InEdge.Backpressure != "dropOldest" || reg.InEdge.Type != "string" || reg.InEdge.UpstreamName != "P" {
        t.Fatalf("unexpected R structured inEdge: %+v", reg.InEdge)
    }
}
