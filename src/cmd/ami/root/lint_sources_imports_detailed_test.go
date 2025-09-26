package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func TestLint_JSON_Sources_ImportsDetailed_IncludesConstraints(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    ws := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
        t.Fatalf("write ws: %v", err)
    }
    if err := os.MkdirAll("src", 0o755); err != nil {
        t.Fatalf("mkdir src: %v", err)
    }
    // Create constrained imports
    src := "package main\nimport ami/stdlib/io >= v0.1.2\nimport (\n  github.com/asymmetric-effort/ami/stdio >= v0.0.0\n  \"fmt\"\n)\n"
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
        t.Fatalf("write src: %v", err)
    }

    // Run lint in JSON mode and capture sources.v1
    old := os.Args
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func() { _ = rootcmd.Execute() })
    os.Args = old

    var sources sch.SourcesV1
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
            continue
        }
        if obj["schema"] == "sources.v1" {
            if err := json.Unmarshal([]byte(sc.Text()), &sources); err != nil {
                t.Fatalf("unmarshal sources: %v", err)
            }
            break
        }
    }
    if sources.Schema != "sources.v1" || len(sources.Units) == 0 {
        t.Fatalf("sources.v1 not found in output")
    }
    // Verify ImportsDetailed contains constraints
    var have = map[string]string{}
    for _, it := range sources.Units[0].ImportsDetailed {
        have[it.Path] = it.Constraint
    }
    if have["ami/stdlib/io"] != ">= v0.1.2" {
        t.Fatalf("unexpected constraint for io: %q", have["ami/stdlib/io"])
    }
    if have["github.com/asymmetric-effort/ami/stdio"] != ">= v0.0.0" {
        t.Fatalf("unexpected constraint for stdio: %q", have["github.com/asymmetric-effort/ami/stdio"])
    }
    if c, ok := have["fmt"]; !ok || c != "" {
        t.Fatalf("fmt should have empty constraint; got %q, ok=%v", c, ok)
    }
}

