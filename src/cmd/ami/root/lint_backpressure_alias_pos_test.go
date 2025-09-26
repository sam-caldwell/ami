package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

func TestLint_JSON_BackpressureAlias_Pos(t *testing.T) {
    ws := t.TempDir()
    content := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatal(err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatal(err) }
    src := "package main\npipeline P { Ingress(cfg).Egress(in=edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=drop,type=string)) }\n"
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatal(err) }
    oldDir, _ := os.Getwd(); _ = os.Chdir(ws); defer os.Chdir(oldDir)
    oldArgs := os.Args
    out := captureStdoutBP(t, func() { os.Args = []string{"ami", "--json", "lint"}; _ = rootcmd.Execute() })
    os.Args = oldArgs
    // Expect alias warning with a position
    seen := false
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var m map[string]any
        if json.Unmarshal([]byte(sc.Text()), &m) != nil { continue }
        if m["schema"] != "diag.v1" || m["code"] != "W_EDGE_BP_ALIAS" { continue }
        if pos, ok := m["pos"].(map[string]any); ok {
            if _, ok2 := pos["line"]; ok2 { seen = true; break }
        }
    }
    if !seen { t.Fatalf("expected W_EDGE_BP_ALIAS with pos; out=\n%s", out) }
}

