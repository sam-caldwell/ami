package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// Ensure E_WORKER_SIGNATURE includes a position in JSON build diagnostics.
func TestBuild_JSON_SemanticDiagnostics_EWorkerSignature_Pos(t *testing.T) {
    ws := t.TempDir()
    wsContent := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatal(err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatal(err) }
    src := "package main\nfunc bad(a int) int {}\npipeline P { Ingress(cfg).Transform(bad).Egress(cfg) }\n"
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatal(err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil { t.Fatalf("expected exit 1 for sem diag; stdout=\n%s", string(out)) }

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_WORKER_SIGNATURE" {
            if pos, ok := obj["pos"].(map[string]any); ok {
                if _, ok2 := pos["line"]; ok2 { seen = true; break }
            }
        }
    }
    if !seen { t.Fatalf("expected E_WORKER_SIGNATURE with pos in JSON output; out=\n%s", string(out)) }
}

