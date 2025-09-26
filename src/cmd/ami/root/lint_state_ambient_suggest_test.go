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

// Ensure lint emits W_STATE_PARAM_AMBIENT_SUGGEST (info) for non-pointer State parameters.
func TestLint_JSON_StateParam_AmbientSuggest_Info(t *testing.T) {
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
    // Non-pointer State parameter
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
        if obj["schema"] == "diag.v1" && obj["code"] == "W_STATE_PARAM_AMBIENT_SUGGEST" {
            if lvl, _ := obj["level"].(string); strings.ToLower(lvl) == "info" { seen = true; break }
        }
    }
    if !seen { t.Fatalf("expected W_STATE_PARAM_AMBIENT_SUGGEST (info) in lint output; out=\n%s", out) }
}

