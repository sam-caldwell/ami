package root_test

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// Ensure edges.v1 captures MultiPath edge entries with normalized inputs/merge.
func TestBuild_Edges_MultiPath_DebugArtifacts(t *testing.T) {
    t.Setenv("AMI_SEM_DIAGS", "0")
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // workspace and source
    wsContent := `version: 1.0.0
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
    if err := os.WriteFile("ami.workspace", []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    src := `package main
func f(ctx Context, ev Event<int>, st State) Event<int> { }
pipeline Up { Ingress(cfg).Transform(f).Egress() }
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=int), edge.Pipeline(name=Up,minCapacity=0,maxCapacity=0,backpressure=dropNewest,type=int)], merge=Sort("ts","asc"))).Egress() }
`
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    // run build --verbose
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "build", "--verbose"}
    _ = captureStdoutBuild(t, func() { _ = rootcmd.Execute() })

    // edges summary JSON present and includes our MultiPath edge with nested fields
    epath := filepath.Join("build", "debug", "asm", "main", "edges.json")
    b, err := os.ReadFile(epath)
    if err != nil { t.Fatalf("missing edges.json: %v", err) }
    var edges sch.EdgesV1
    if err := json.Unmarshal(b, &edges); err != nil { t.Fatalf("unmarshal edges: %v", err) }
    if err := edges.Validate(); err != nil { t.Fatalf("edges validate: %v", err) }
    var ok bool
    for _, it := range edges.Items {
        if it.Label == "P.step1.in" && it.Kind == "edge.MultiPath" && it.MultiPath != nil {
            if len(it.MultiPath.Inputs) != 2 { t.Fatalf("expected 2 inputs; got %d", len(it.MultiPath.Inputs)) }
            if len(it.MultiPath.Merge) == 0 || it.MultiPath.Merge[0].Name == "" { t.Fatalf("expected merge op; got %+v", it.MultiPath.Merge) }
            ok = true
            break
        }
    }
    if !ok { t.Fatalf("edges.json missing MultiPath entry with details; items=%+v", edges.Items) }
}

