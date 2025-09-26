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

func TestLint_JSON_MultiPath_Hints(t *testing.T) {
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
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    src := `package main
pipeline P { Ingress(cfg).Transform(worker=w, in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=0,maxCapacity=0,backpressure=block,type=int)])).Egress() }
func w(ctx Context, ev Event<int>, st State) Event<int> { }`
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    // Run lint JSON
    old := os.Args
    defer func(){ os.Args = old }()
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    // scan for MultiPath warnings
    has := map[string]bool{}
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" { continue }
        var m map[string]any
        if json.Unmarshal([]byte(line), &m) != nil { continue }
        if m["schema"] == "diag.v1" {
            code, _ := m["code"].(string)
            if strings.HasPrefix(code, "W_MP_") { has[code] = true }
        }
    }
    if !has["W_MP_EDGE_SMELL_UNBOUNDED_BLOCK"] { t.Fatalf("expected W_MP_EDGE_SMELL_UNBOUNDED_BLOCK; got %v", has) }
}
