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

// Reuse TestHelper_AmiBuildJSON

func TestBuild_JSON_SemanticDiagnostics_EWorkerSignature(t *testing.T) {
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
    src := `package main
func bad(Context, Event<string>) Event<string> {}
pipeline P { Ingress(cfg).Transform(bad).Egress(cfg) }
`
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatal(err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil { t.Fatalf("expected exit 1 for sem diag; stdout=\n%s", string(out)) }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 1 { t.Fatalf("got exit %d want 1; out=\n%s", code, string(out)) }
    } else { t.Fatalf("unexpected err type: %T", err) }

    // Find E_WORKER_SIGNATURE diag
    var seen bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_WORKER_SIGNATURE" {
            seen = true
            break
        }
    }
    if !seen { t.Fatalf("expected E_WORKER_SIGNATURE in JSON output; out=\n%s", string(out)) }
}

