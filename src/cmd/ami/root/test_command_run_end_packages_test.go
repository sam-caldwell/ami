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

// Verify run_end includes per-package summaries
func TestTest_JSON_RunEndIncludesPerPackageSummaries(t *testing.T) {
    ws := t.TempDir()
    gomod := "module example.com/ami-test\n\ngo 1.22\n"
    if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil { t.Fatalf("write go.mod: %v", err) }
    // Create p1 and p2 each with one passing test
    if err := os.MkdirAll(filepath.Join(ws, "p1"), 0o755); err != nil { t.Fatalf("mkdir p1: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "p2"), 0o755); err != nil { t.Fatalf("mkdir p2: %v", err) }
    src1 := `package p1
import "testing"
func TestOne(t *testing.T){ }
`
    src2 := `package p2
import "testing"
func TestTwo(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(ws, "p1", "a_test.go"), []byte(src1), 0o644); err != nil { t.Fatalf("write p1: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "p2", "a_test.go"), []byte(src2), 0o644); err != nil { t.Fatalf("write p2: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected exit: %v; stdout=\n%s", err, string(out)) }

    type pkgSummary struct{ Package string; Pass, Fail, Skip, Cases int }
    type runEnd struct {
        Schema string `json:"schema"`
        Type   string `json:"type"`
        Packages []pkgSummary `json:"packages"`
    }
    var seen runEnd
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var probe map[string]any
        if json.Unmarshal([]byte(sc.Text()), &probe) != nil { continue }
        if probe["schema"] == "test.v1" && probe["type"] == "run_end" {
            if json.Unmarshal([]byte(sc.Text()), &seen) == nil { break }
        }
    }
    if len(seen.Packages) < 2 { t.Fatalf("expected >=2 package summaries; got %d", len(seen.Packages)) }
    // Expect entries for p1 and p2
    want1 := "example.com/ami-test/p1"
    want2 := "example.com/ami-test/p2"
    var ok1, ok2 bool
    for _, p := range seen.Packages {
        if p.Package == want1 && p.Cases == 1 && p.Pass == 1 && p.Fail == 0 && p.Skip == 0 { ok1 = true }
        if p.Package == want2 && p.Cases == 1 && p.Pass == 1 && p.Fail == 0 && p.Skip == 0 { ok2 = true }
    }
    if !ok1 || !ok2 {
        t.Fatalf("did not observe expected package summaries; got: %#v", seen.Packages)
    }
}

