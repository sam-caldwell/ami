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

// Ensures runtime cases are discovered and currently skipped by the stub harness.
func TestTest_JSON_AMIRuntime_JsonEqualityPass(t *testing.T) {
    ws := t.TempDir()
    gomod := "module example.com/ami-test\n\ngo 1.22\n"
    if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil { t.Fatalf("write go.mod: %v", err) }
    wsContent := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }
packages: [ { main: { version: 0.0.1, root: ./src, import: [] } } ]
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    // Runtime test case pragma; identity behavior yields pass.
    a := `#pragma test:case RunEcho
#pragma test:runtime pipeline=P input={"x":1} expect_output={"x":1}
package main
Pipeline(P) { }
`
    if err := os.WriteFile(filepath.Join(ws, "src", "rt1_test.ami"), []byte(a), 0o644); err != nil { t.Fatalf("write rt1: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil {
        if ee, ok := err.(*exec.ExitError); ok { _ = ee } else { t.Fatalf("unexpected error: %v\n%s", err, string(out)) }
    }
    // Look for test_end event with status=pass for RunEcho
    sawPass := false
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "test.v1" && obj["type"] == "test_end" && obj["status"] == "pass" {
            if name, _ := obj["name"].(string); name == "RunEcho" { sawPass = true }
        }
    }
    if !sawPass { t.Fatalf("expected pass for runtime case; stdout=\n%s", string(out)) }
}
