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

// Reuse TestHelper_AmiTestJSON helper

func TestTest_JSON_SystemIOErrorOnBuildFailure(t *testing.T) {
    ws := t.TempDir()
    // invalid code to cause build failure (syntax error)
    gomod := "module example.com/ami-test\n\ngo 1.22\n"
    if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil { t.Fatalf("write go.mod: %v", err) }
    bad := `package bad
func Oops( { }
`
    if err := os.WriteFile(filepath.Join(ws, "oops.go"), []byte(bad), 0o644); err != nil { t.Fatalf("write bad: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil {
        t.Fatalf("expected non-zero exit for build error; stdout=\n%s", string(out))
    }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 2 {
            t.Fatalf("unexpected exit code: got %d want 2; stdout=\n%s", code, string(out))
        }
    } else {
        t.Fatalf("unexpected error type: %T", err)
    }

    // Ensure we still saw a run_start/run_end envelope even with failure
    var sawStart, sawEnd bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var rec struct{ Schema, Type string }
        if json.Unmarshal([]byte(sc.Text()), &rec) != nil { continue }
        if rec.Schema == "test.v1" && rec.Type == "run_start" { sawStart = true }
        if rec.Schema == "test.v1" && rec.Type == "run_end" { sawEnd = true }
    }
    if !sawStart || !sawEnd { t.Fatalf("expected run_start/run_end in JSON output; got:\n%s", string(out)) }
}

