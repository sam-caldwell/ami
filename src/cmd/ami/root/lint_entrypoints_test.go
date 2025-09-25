package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    sch "github.com/sam-caldwell/ami/src/schemas"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

// captureStdoutLint is scoped for lint tests
func captureStdoutLint(t *testing.T, fn func()) string {
    t.Helper()
    old := os.Stdout
    r, w, err := os.Pipe()
    if err != nil { t.Fatalf("pipe: %v", err) }
    os.Stdout = w
    fn()
    _ = w.Close()
    os.Stdout = old
    var b strings.Builder
    sc := bufio.NewScanner(r)
    for sc.Scan() { b.WriteString(sc.Text()); b.WriteByte('\n') }
    return b.String()
}

func TestLint_JSON_EntryOrder_ImportsBeforeMain(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // Workspace with main and a local package 'lib/util'
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
  - lib/util:
      version: 0.0.1
      root: ./lib/util
      import: []
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // create sources
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.MkdirAll(filepath.Join("lib","util"), 0o755); err != nil { t.Fatalf("mkdir lib/util: %v", err) }
    // lib util unit
    if err := os.WriteFile(filepath.Join("lib","util","util.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatalf("write util: %v", err) }
    // main imports lib/util and an external fmt
    mainSrc := "package main\nimport \"lib/util\"\nimport \"fmt\"\n"
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(mainSrc), 0o644); err != nil { t.Fatalf("write main: %v", err) }

    // Run ami --json lint
    oldArgs := os.Args
    out := captureStdoutLint(t, func(){
        os.Args = []string{"ami", "--json", "lint"}
        _ = rootcmd.Execute()
    })
    os.Args = oldArgs

    // Find a sources.v1 object
    var seen sch.SourcesV1
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" { continue }
        var probe map[string]any
        if json.Unmarshal([]byte(line), &probe) != nil { continue }
        if probe["schema"] == "sources.v1" {
            var s sch.SourcesV1
            if json.Unmarshal([]byte(line), &s) == nil && s.Validate() == nil {
                seen = s
                break
            }
        }
    }
    if len(seen.Units) < 2 { t.Fatalf("expected at least two units (imported + main)") }
    // The first unit should belong to lib/util, followed by main
    if !(strings.HasSuffix(seen.Units[0].File, "lib/util/util.ami") && strings.HasSuffix(seen.Units[len(seen.Units)-1].File, "src/main.ami")) {
        t.Fatalf("unexpected unit order: first=%s last=%s", seen.Units[0].File, seen.Units[len(seen.Units)-1].File)
    }
}

func TestLint_Human_ListsUnitsInOrder(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

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
  - lib/util:
      version: 0.0.1
      root: ./lib/util
      import: []
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.MkdirAll(filepath.Join("lib","util"), 0o755); err != nil { t.Fatalf("mkdir lib/util: %v", err) }
    if err := os.WriteFile(filepath.Join("lib","util","util.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatalf("write util: %v", err) }
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\nimport \"lib/util\"\n"), 0o644); err != nil { t.Fatalf("write main: %v", err) }

    // Run human mode
    oldArgs := os.Args
    os.Args = []string{"ami", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    os.Args = oldArgs

    // Assert the lib/util unit appears before main unit in output
    idxLib := strings.Index(out, "lib/util/util.ami")
    idxMain := strings.Index(out, "src/main.ami")
    if idxLib == -1 || idxMain == -1 || !(idxLib < idxMain) {
        t.Fatalf("expected lib/util before main in output; got:\n%s", out)
    }
}
