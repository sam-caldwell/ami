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

// normalizeJSON parses known debug artifact JSON files into schema structs
// and zeroes the top-level Timestamp field (allowed to vary) for comparison.
func normalizeJSON(t *testing.T, path string, b []byte) any {
    t.Helper()
    switch {
    case filepath.Base(path) == "resolved.json":
        var v sch.SourcesV1
        if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal sources: %v", err) }
        v.Timestamp = ""
        return v
    case filepath.Ext(path) == ".json" && filepath.Base(path) == "index.json":
        var v sch.ASMIndexV1
        if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal asm index: %v", err) }
        v.Timestamp = ""
        return v
    case filepath.Ext(path) == ".json" && filepath.Base(path) == "edges.json":
        var v sch.EdgesV1
        if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal edges: %v", err) }
        v.Timestamp = ""
        return v
    case filepath.Ext(path) == ".json" && filepath.Ext(filepath.Base(path[:len(path)-len(filepath.Ext(path))])) == ".ast":
        var v sch.ASTV1
        if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal ast: %v", err) }
        v.Timestamp = ""
        return v
    case filepath.Ext(path) == ".json" && filepath.Ext(filepath.Base(path[:len(path)-len(filepath.Ext(path))])) == ".ir":
        var v sch.IRV1
        if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal ir: %v", err) }
        v.Timestamp = ""
        return v
    case filepath.Ext(path) == ".json" && filepath.Ext(filepath.Base(path[:len(path)-len(filepath.Ext(path))])) == ".pipelines":
        var v sch.PipelinesV1
        if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal pipelines: %v", err) }
        v.Timestamp = ""
        return v
    case filepath.Ext(path) == ".json" && filepath.Ext(filepath.Base(path[:len(path)-len(filepath.Ext(path))])) == ".eventmeta":
        var v sch.EventMetaV1
        if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal eventmeta: %v", err) }
        v.Timestamp = ""
        return v
    default:
        t.Fatalf("unexpected json artifact path: %s", path)
        return nil
    }
}

func TestBuild_DebugArtifacts_DeterministicAcrossRuns(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // disable sem diags to focus on artifact determinism only
    t.Setenv("AMI_SEM_DIAGS", "0")
    // minimal workspace + source with a pipeline and worker
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
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    src := `package main
func w(ctx Context, ev Event<string>, st State) Event<string> { }
pipeline P { Ingress(cfg).Transform(w).Egress(in=edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=string)) }
`
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    // Run build --verbose twice
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "build", "--verbose"}
    _ = rootcmd.Execute()

    // capture artifacts from first run
    paths := []string{
        filepath.Join("build","debug","source","resolved.json"),
        filepath.Join("build","debug","ast","main","main.ami.ast.json"),
        filepath.Join("build","debug","ir","main","main.ami.ir.json"),
        filepath.Join("build","debug","ir","main","main.ami.pipelines.json"),
        filepath.Join("build","debug","ir","main","main.ami.eventmeta.json"),
        filepath.Join("build","debug","asm","main","main.ami.s"),
        filepath.Join("build","debug","asm","main","index.json"),
        filepath.Join("build","debug","asm","main","edges.json"),
    }

    type snapshot struct {
        raw []byte
        norm any
    }
    snap1 := map[string]snapshot{}
    for _, p := range paths {
        b, err := os.ReadFile(p)
        if err != nil { t.Fatalf("missing artifact %s: %v", p, err) }
        if filepath.Ext(p) == ".s" {
            snap1[p] = snapshot{raw: b}
            continue
        }
        snap1[p] = snapshot{raw: b, norm: normalizeJSON(t, p, b)}
    }

    // second build
    _ = rootcmd.Execute()

    // compare artifacts
    for _, p := range paths {
        b2, err := os.ReadFile(p)
        if err != nil { t.Fatalf("artifact missing after second build %s: %v", p, err) }
        first := snap1[p]
        if filepath.Ext(p) == ".s" {
            if !reflect.DeepEqual(first.raw, b2) {
                t.Fatalf("asm content changed across runs for %s", p)
            }
            continue
        }
        norm2 := normalizeJSON(t, p, b2)
        if !reflect.DeepEqual(first.norm, norm2) {
            t.Fatalf("json artifact changed across runs for %s", p)
        }
    }
}
