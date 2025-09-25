package root_test

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    sch "github.com/sam-caldwell/ami/src/schemas"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

// Verifies that build --verbose emits a .pipelines.json file and that it
// validates against the pipelines.v1 schema.
func TestBuild_PipelinesIR_EmittedAndValid(t *testing.T) {
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
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatal(err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatal(err) }
    // A simple worker and pipeline so pipelines.json is non-empty
src := `package main
func x(ctx Context, ev Event<T>, st State) Event<U> { }
pipeline P { Ingress(cfg).Transform(x).Egress(cfg) }
`
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(src), 0o644); err != nil { t.Fatal(err) }

    // Run build --verbose
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "build", "--verbose"}
    _ = captureStdoutBuild(t, func(){ _ = rootcmd.Execute() })

    // Expect pipelines JSON
    path := filepath.Join("build","debug","ir","main","main.ami.pipelines.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("missing pipelines json: %v", err) }
    var p sch.PipelinesV1
    if err := json.Unmarshal(b, &p); err != nil { t.Fatalf("unmarshal pipelines: %v", err) }
    if err := p.Validate(); err != nil { t.Fatalf("pipelines schema validate: %v", err) }
    if len(p.Pipelines) == 0 { t.Fatalf("expected at least one pipeline record") }
    // find pipeline P
    var pipe *sch.PipelineV1
    for i := range p.Pipelines { if p.Pipelines[i].Name == "P" { pipe = &p.Pipelines[i]; break } }
    if pipe == nil { t.Fatalf("expected pipeline named P in pipelines.json; got %+v", p.Pipelines) }
    if len(pipe.Steps) == 0 { t.Fatalf("expected at least one step in pipeline P") }
    // ensure worker x is referenced
    foundWorker := false
    var worker sch.PipelineWorkerV1
    for _, st := range pipe.Steps {
        for _, w := range st.Workers { if w.Name == "x" { foundWorker = true; worker = w; break } }
        if foundWorker { break }
    }
    if !foundWorker { t.Fatalf("expected to find worker 'x' in pipeline P steps") }
    if worker.OutputKind == "" { t.Fatalf("expected worker 'x' to have outputKind populated") }
    if worker.Input == "" { t.Fatalf("expected worker 'x' to have input payload type populated") }
    if worker.Output == "" { t.Fatalf("expected worker 'x' to have output payload type populated") }
}
