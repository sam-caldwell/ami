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

// Ensure workspace severity config applies to parser errors as well.
func TestLint_Config_ParserErrorSeverityOverride(t *testing.T) {
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
  linter:
    rules:
      E_BAD_PACKAGE: info
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    // Invalid package identifier → E_BAD_PACKAGE from parser
    src := "package 123\n"
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    // Run CLI JSON
    old := os.Args
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    os.Args = old

    var ok bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_BAD_PACKAGE" {
            if lvl, _ := obj["level"].(string); lvl == "info" { ok = true; break }
        }
    }
    if !ok { t.Fatalf("expected E_BAD_PACKAGE with level info; output:\n%s", out) }
}

