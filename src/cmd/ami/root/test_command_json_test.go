package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// Helper that runs `ami --json test ./...` in this process when enabled.
func TestHelper_AmiTestJSON(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI_TEST_JSON") != "1" { return }
    os.Args = []string{"ami", "--json", "test", "./..."}
    code := rootcmd.Execute()
    os.Exit(code)
}

// Helper variant without args (defaults to ./...)
func TestHelper_AmiTestJSON_NoArgs(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI_TEST_JSON_NOARGS") != "1" { return }
    os.Args = []string{"ami", "--json", "test"}
    code := rootcmd.Execute()
    os.Exit(code)
}

// Helper with timeout flag
func TestHelper_AmiTestJSON_Timeout50ms(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI_TEST_JSON_TIMEOUT") != "1" { return }
    os.Args = []string{"ami", "--json", "test", "--timeout", "50ms", "./..."}
    code := rootcmd.Execute()
    os.Exit(code)
}

// Helper with parallel flag
func TestHelper_AmiTestJSON_Parallel1(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI_TEST_JSON_PARALLEL") != "1" { return }
    os.Args = []string{"ami", "--json", "test", "--parallel", "1", "./..."}
    code := rootcmd.Execute()
    os.Exit(code)
}

// Helper variant for pkg-parallel
func TestHelper_AmiTestJSON_ParallelPkg2(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI_TEST_JSON_PKGPAR") != "1" { return }
    os.Args = []string{"ami", "--json", "test", "--pkg-parallel", "2", "./..."}
    os.Setenv("AMI_TEST_PKG_PARALLEL", "2")
    code := rootcmd.Execute()
    os.Exit(code)
}

// Minimal test.v1 record probe
type testRecord struct {
    Schema string `json:"schema"`
    Type   string `json:"type"`
}

func TestTest_JSON_SimplePass(t *testing.T) {
    ws := t.TempDir()
    // Create a minimal go module with one passing test
    gomod := "module example.com/ami-test\n\ngo 1.22\n"
    if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil { t.Fatalf("write go.mod: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "pkg"), 0o755); err != nil { t.Fatalf("mkdir pkg: %v", err) }
    testSrc := `package pkg
import "testing"
func TestHello(t *testing.T){ t.Log("hello"); if 1!=1 { t.Fatal("bad math") } }
`
    if err := os.WriteFile(filepath.Join(ws, "pkg", "a_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write test: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("unexpected non-zero exit running ami test: %v\nstdout=\n%s", err, string(out))
    }

    // Validate that we observe test.v1 events (run_start and run_end at minimum)
    var sawStart, sawEnd bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var rec testRecord
        if json.Unmarshal([]byte(sc.Text()), &rec) != nil { continue }
        if rec.Schema == "test.v1" && rec.Type == "run_start" { sawStart = true }
        if rec.Schema == "test.v1" && rec.Type == "run_end" { sawEnd = true }
    }
    if !sawStart || !sawEnd {
        t.Fatalf("did not observe expected test.v1 start/end events. stdout=\n%s", string(out))
    }
}
