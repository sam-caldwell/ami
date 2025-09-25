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

// Validates that AMI native test cases (in *_test.ami) are discovered and executed,
// covering pass/fail/skip paths.
func TestTest_JSON_RunAmiNativeCases(t *testing.T) {
    ws := t.TempDir()
    // minimal go module so `go test` runs quickly
    gomod := "module example.com/ami-test\n\ngo 1.22\n"
    if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil { t.Fatalf("write go.mod: %v", err) }
    // Workspace for AMI CLI
    wsContent := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }
packages: [ { main: { version: 0.0.1, root: ./src, import: [] } } ]
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    // Case 1: pass (no errors)
    a := "#pragma test:case pass_case\n#pragma test:expect_no_errors\npackage main\n"
    if err := os.WriteFile(filepath.Join(ws, "src", "a_test.ami"), []byte(a), 0o644); err != nil { t.Fatalf("write a_test: %v", err) }
    // Case 2: fail (mismatch expectation)
    b := "#pragma test:case fail_case\n#pragma test:expect_error E_NOT_A_CODE\npackage main\n"
    if err := os.WriteFile(filepath.Join(ws, "src", "b_test.ami"), []byte(b), 0o644); err != nil { t.Fatalf("write b_test: %v", err) }
    // Case 3: skip
    c := "#pragma test:case skip_case\n#pragma test:skip not_implemented\npackage main\n"
    if err := os.WriteFile(filepath.Join(ws, "src", "c_test.ami"), []byte(c), 0o644); err != nil { t.Fatalf("write c_test: %v", err) }

    // Run ami --json test
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil {
        if ee, ok := err.(*exec.ExitError); ok {
            if code := ee.ExitCode(); code != 1 { t.Fatalf("unexpected exit code: %d; stdout=\n%s", code, string(out)) }
        } else {
            t.Fatalf("unexpected error type: %T; stdout=\n%s", err, string(out))
        }
    }

    // Count test_end statuses for the three AMI cases
    var pass, fail, skip int
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "test.v1" && obj["type"] == "test_end" {
            switch obj["status"] {
            case "pass": pass++
            case "fail": fail++
            case "skip": skip++
            }
        }
    }
    if pass < 1 || fail < 1 || skip < 1 {
        t.Fatalf("expected at least one pass, fail, and skip; got pass=%d fail=%d skip=%d\n%s", pass, fail, skip, string(out))
    }
}
