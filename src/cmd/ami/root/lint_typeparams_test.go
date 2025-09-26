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

// Assert that duplicate function type parameters are reported by `ami --json lint`
// as E_DUP_TYPE_PARAM.
func TestLint_JSON_DuplicateTypeParams(t *testing.T) {
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
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
        t.Fatalf("write ws: %v", err)
    }
    if err := os.MkdirAll("src", 0o755); err != nil {
        t.Fatalf("mkdir src: %v", err)
    }
    // Duplicate type parameter names: func F<T, T>(a T) {}
    src := "package main\nfunc F<T, T>(a T) {}\n"
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
        t.Fatalf("write src: %v", err)
    }

    oldArgs := os.Args
    out := captureStdoutLintRules(t, func() {
        os.Args = []string{"ami", "--json", "lint"}
        _ = rootcmd.Execute()
    })
    os.Args = oldArgs

    // Look for diag.v1 with code E_DUP_TYPE_PARAM
    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
            continue
        }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_DUP_TYPE_PARAM" {
            // basic position presence check
            if pos, ok := obj["pos"].(map[string]any); ok {
                if _, ok2 := pos["line"]; ok2 {
                    seen = true
                    break
                }
            }
        }
    }
    if !seen {
        t.Fatalf("expected E_DUP_TYPE_PARAM in lint output; got:\n%s", out)
    }
}
