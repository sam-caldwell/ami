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

// Assert an unbounded edge (maxCapacity=0) emits bounded=false and delivery=bestEffort (backpressure=drop).
func TestBuild_Edges_UnboundedBestEffort(t *testing.T) {
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
func f(ctx Context, ev Event<string>, st State) Event<string> { }
pipeline P { Ingress(cfg).Transform(f).Egress(in=edge.FIFO(minCapacity=0,maxCapacity=0,backpressure=drop,type=string)) }
`
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    // run build --verbose
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "build", "--verbose"}
    _ = captureStdoutBuild(t, func(){ _ = rootcmd.Execute() })

    // load edges.json
    epath := filepath.Join("build","debug","asm","main","edges.json")
    b, err := os.ReadFile(epath)
    if err != nil { t.Fatalf("missing edges.json: %v", err) }
    var edges sch.EdgesV1
    if err := json.Unmarshal(b, &edges); err != nil { t.Fatalf("unmarshal edges: %v", err) }
    if err := edges.Validate(); err != nil { t.Fatalf("edges validate: %v", err) }
    // find our edge and assert bounded=false, delivery=bestEffort
    var ok bool
    for _, it := range edges.Items {
        if it.Label == "P.step2.in" && it.Kind == "edge.FIFO" {
            if it.Bounded { t.Fatalf("expected bounded=false; got true") }
            if it.Delivery != "bestEffort" { t.Fatalf("expected delivery=bestEffort; got %s", it.Delivery) }
            ok = true
            break
        }
    }
    if !ok { t.Fatalf("expected FIFO edge P.step2.in with bounded=false bestEffort in edges.json; got: %+v", edges.Items) }
}
