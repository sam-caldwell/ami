package root_test

import (
    "bufio"
    "encoding/json"
    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Assert E_WORKER_SIGNATURE includes position in lint JSON mode.
func TestLint_JSON_WorkerSignature_Pos(t *testing.T) {
    t.Setenv("HOME", t.TempDir())
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
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatal(err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatal(err) }
    // Invalid worker signature: func bad(a int) int
    src := "package main\nfunc bad(a int) int {}\npipeline P { Ingress(cfg).Transform(bad).Egress(cfg) }\n"
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatal(err) }

    oldArgs := os.Args
    out := captureStdoutLintRules(t, func() { os.Args = []string{"ami", "--json", "lint"}; _ = rootcmd.Execute() })
    os.Args = oldArgs

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_WORKER_SIGNATURE" {
            if pos, ok := obj["pos"].(map[string]any); ok {
                if _, ok2 := pos["line"]; ok2 { seen = true; break }
            }
        }
    }
    if !seen { t.Fatalf("expected E_WORKER_SIGNATURE with pos in lint output; out=\n%s", out) }
}

