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

// Exercises count and msg substring expectations via CLI.
func TestTest_JSON_AMIAssertions_CountAndSubstring(t *testing.T) {
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
    // Two imports → two E_BAD_IMPORT
    countCase := `#pragma test:case CountOk
#pragma test:expect_error E_BAD_IMPORT count=2
package main
import "bad path"
import "bad path"
`
    if err := os.WriteFile(filepath.Join(ws, "src", "count_test.ami"), []byte(countCase), 0o644); err != nil { t.Fatalf("write count: %v", err) }
    substrCase := `#pragma test:case SubstrOk
#pragma test:expect_error E_BAD_IMPORT count=1 msg~="invalid import"
package main
import "bad path"
`
    if err := os.WriteFile(filepath.Join(ws, "src", "substr_test.ami"), []byte(substrCase), 0o644); err != nil { t.Fatalf("write substr: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil {
        if ee, ok := err.(*exec.ExitError); ok { // may exit 1 if any other failures present
            _ = ee
        } else { t.Fatalf("unexpected error: %v\n%s", err, string(out)) }
    }
    // Verify both cases ended with status=pass
    want := map[string]bool{"CountOk": false, "SubstrOk": false}
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "test.v1" && obj["type"] == "test_end" && obj["status"] == "pass" {
            if name, _ := obj["name"].(string); want[name] == false { want[name] = true }
        }
    }
    for k, v := range want { if !v { t.Fatalf("expected pass for %s; output=\n%s", k, string(out)) } }
}

