package root_test

import (
    "encoding/json"
    "os"
    "path/filepath"
    "reflect"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    sch "github.com/sam-caldwell/ami/src/schemas"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestBuild_ObjIndex_DeterministicAcrossRuns(t *testing.T) {
    t.Setenv("HOME", t.TempDir())
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // workspace + trivial source
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
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\n"), 0o644); err != nil { t.Fatal(err) }

    // first build
    old := os.Args
    os.Args = []string{"ami", "build"}
    _ = rootcmd.Execute()

    // snapshot index and obj file
    idxPath := filepath.Join("build","obj","main","index.json")
    asmPath := filepath.Join("build","obj","main","main.ami.s")
    idx1b, err := os.ReadFile(idxPath)
    if err != nil { t.Fatalf("read index1: %v", err) }
    asm1b, err := os.ReadFile(asmPath)
    if err != nil { t.Fatalf("read asm1: %v", err) }
    var idx1 sch.ObjIndexV1
    if err := json.Unmarshal(idx1b, &idx1); err != nil { t.Fatalf("unmarshal idx1: %v", err) }
    // normalize timestamp
    idx1.Timestamp = ""

    // second build
    _ = rootcmd.Execute()
    os.Args = old

    idx2b, err := os.ReadFile(idxPath)
    if err != nil { t.Fatalf("read index2: %v", err) }
    asm2b, err := os.ReadFile(asmPath)
    if err != nil { t.Fatalf("read asm2: %v", err) }
    var idx2 sch.ObjIndexV1
    if err := json.Unmarshal(idx2b, &idx2); err != nil { t.Fatalf("unmarshal idx2: %v", err) }
    idx2.Timestamp = ""

    if !reflect.DeepEqual(idx1, idx2) {
        t.Fatalf("obj index changed across runs:\n1=%s\n2=%s", string(idx1b), string(idx2b))
    }
    if !reflect.DeepEqual(asm1b, asm2b) {
        t.Fatalf("obj asm changed across runs")
    }
}

