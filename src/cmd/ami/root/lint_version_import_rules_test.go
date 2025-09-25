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

func captureStdoutLintPkgRules(t *testing.T, fn func()) string {
    t.Helper()
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    fn()
    _ = w.Close()
    os.Stdout = old
    var b strings.Builder
    sc := bufio.NewScanner(r)
    for sc.Scan() { b.WriteString(sc.Text()); b.WriteByte('\n') }
    return b.String()
}

func TestLint_JSON_WorkspacePackageVersion_Semver(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // invalid package version should emit E_WS_PKG_VERSION
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
  - bad/pkg:
      version: not-semver
      root: ./pkg
      import: []
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    _ = os.MkdirAll("src", 0o755)
    _ = os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\n"), 0o644)

    oldArgs := os.Args
    out := captureStdoutLintPkgRules(t, func(){
        os.Args = []string{"ami", "--json", "lint"}
        _ = rootcmd.Execute()
    })
    os.Args = oldArgs

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_WS_PKG_VERSION" {
            seen = true
            break
        }
    }
    if !seen { t.Fatalf("expected E_WS_PKG_VERSION in lint output; got:\n%s", out) }
}

func TestLint_JSON_WorkspacePackageName_Invalid(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // invalid package name (contains '+') should emit E_WS_PKG_NAME
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
  - lib+util:
      version: 0.0.1
      root: ./lib/util
      import: []
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    _ = os.MkdirAll("src", 0o755)
    _ = os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\n"), 0o644)

    oldArgs := os.Args
    out := captureStdoutLintPkgRules(t, func(){
        os.Args = []string{"ami", "--json", "lint"}
        _ = rootcmd.Execute()
    })
    os.Args = oldArgs

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_WS_PKG_NAME" {
            seen = true
            break
        }
    }
    if !seen { t.Fatalf("expected E_WS_PKG_NAME in lint output; got:\n%s", out) }
}

