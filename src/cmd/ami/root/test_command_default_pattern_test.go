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

// Reuse TestHelper_AmiTestJSON

func TestTest_JSON_DefaultPatternIsDotDotDot(t *testing.T) {
    ws := t.TempDir()
    gomod := "module example.com/ami-test\n\ngo 1.22\n"
    if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil { t.Fatalf("write go.mod: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "a_test.go"), []byte(`package main
import "testing"
func TestOk(t *testing.T){ }
`), 0o644); err != nil { t.Fatalf("write test: %v", err) }

    // helper defaults to ./... when no args provided
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON_NoArgs")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON_NOARGS=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit code: %v; stdout=\n%s", err, string(out)) }

    // Parse first test.v1 run_start and assert packages contains ./...
    type runStart struct{ Schema, Type string; Packages []string }
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    var seen runStart
    for sc.Scan() {
        var rs runStart
        if json.Unmarshal([]byte(sc.Text()), &rs) != nil { continue }
        if rs.Schema == "test.v1" && rs.Type == "run_start" { seen = rs; break }
    }
    if len(seen.Packages) != 1 || seen.Packages[0] != "./..." {
        t.Fatalf("expected default packages ['./...']; got %#v", seen.Packages)
    }
}
