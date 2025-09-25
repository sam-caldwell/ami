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
)

// Verify prerelease imports are flagged when constraints omit prereleases.
func TestLint_CrossPackage_PrereleaseForbidden(t *testing.T) {
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
  - main: { version: 0.0.1, root: ./src, import: ["lib/util >=0.1.0"] }
  - lib/util: { version: 0.1.0-rc1, root: ./lib/util, import: [] }
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.MkdirAll(filepath.Join("lib","util"), 0o755); err != nil { t.Fatalf("mkdir lib/util: %v", err) }
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\nimport \"lib/util\"\n"), 0o644); err != nil { t.Fatalf("write main: %v", err) }
    if err := os.WriteFile(filepath.Join("lib","util","util.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatalf("write util: %v", err) }

    old := os.Args
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    os.Args = old

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_IMPORT_PRERELEASE_FORBIDDEN" { seen = true; break }
    }
    if !seen { t.Fatalf("expected E_IMPORT_PRERELEASE_FORBIDDEN in output; got:\n%s", out) }
}

// Verify consistency rule flags conflicting constraints across importers.
func TestLint_CrossPackage_ConsistencyRule(t *testing.T) {
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
  - main: { version: 0.0.1, root: ./src, import: ["lib/util ^0.1.0"] }
  - app/other: { version: 0.0.1, root: ./app/other, import: ["lib/util >=0.1.0"] }
  - lib/util: { version: 0.1.3, root: ./lib/util, import: [] }
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.MkdirAll(filepath.Join("app","other"), 0o755); err != nil { t.Fatalf("mkdir app/other: %v", err) }
    if err := os.MkdirAll(filepath.Join("lib","util"), 0o755); err != nil { t.Fatalf("mkdir lib/util: %v", err) }
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\nimport \"lib/util\"\n"), 0o644); err != nil { t.Fatalf("write main: %v", err) }
    if err := os.WriteFile(filepath.Join("app","other","main.ami"), []byte("package other\nimport \"lib/util\"\n"), 0o644); err != nil { t.Fatalf("write other: %v", err) }
    if err := os.WriteFile(filepath.Join("lib","util","util.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatalf("write util: %v", err) }

    old := os.Args
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    os.Args = old

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_IMPORT_CONSISTENCY" { seen = true; break }
    }
    if !seen { t.Fatalf("expected E_IMPORT_CONSISTENCY in output; got:\n%s", out) }
}

