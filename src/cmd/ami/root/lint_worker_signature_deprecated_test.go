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

// Assert legacy 3-parameter worker signature is rejected with E_WORKER_SIGNATURE in lint JSON.
func TestLint_JSON_WorkerSignature_Legacy_Rejected(t *testing.T) {
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
    // Legacy 3-param worker signature (Context, Event<T>, State) — should be rejected
    src := "package main\nfunc f(ctx Context, ev Event<string>, st State) Event<string> { ev }\npipeline P { Ingress(cfg).Transform(f).Egress(cfg) }\n"
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatal(err) }

    oldArgs := os.Args
    out := captureStdoutLintRules(t, func() { os.Args = []string{"ami", "--json", "lint"}; _ = rootcmd.Execute() })
    os.Args = oldArgs

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_WORKER_SIGNATURE" { seen = true; break }
    }
    if !seen { t.Fatalf("expected E_WORKER_SIGNATURE in lint output; out=\n%s", out) }
}
