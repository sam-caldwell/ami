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

func TestBuild_ResolvedSources_ImportsDetailed_IncludesConstraints(t *testing.T) {
    t.Setenv("AMI_SEM_DIAGS", "0")
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // workspace and constrained source
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
    src := "package main\nimport ami/stdlib/io >= v0.1.2\nimport (\n  github.com/asymmetric-effort/ami/stdio >= v0.0.0\n  \"fmt\"\n)\n"
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
        t.Fatalf("write src: %v", err)
    }

    // run build --verbose to emit resolved sources
    old := os.Args
    defer func() { os.Args = old }()
    os.Args = []string{"ami", "build", "--verbose"}
    _ = captureStdoutBuild(t, func() { _ = rootcmd.Execute() })

    // load and validate resolved sources
    b, err := os.ReadFile("build/debug/source/resolved.json")
    if err != nil {
        t.Fatalf("missing resolved.json: %v", err)
    }
    var sources sch.SourcesV1
    if err := json.Unmarshal(b, &sources); err != nil {
        t.Fatalf("unmarshal sources: %v", err)
    }
    if err := sources.Validate(); err != nil {
        t.Fatalf("sources validate: %v", err)
    }
    if len(sources.Units) == 0 {
        t.Fatalf("no source units recorded")
    }
    // verify constraints included
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

