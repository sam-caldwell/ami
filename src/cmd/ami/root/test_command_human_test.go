package root_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// Helper to run ami test in human mode
func TestHelper_AmiTestHuman(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_TEST_HUMAN") != "1" {
		return
	}
	// human mode (no --json)
	os.Args = []string{"ami", "test", "./..."}
	code := rootcmd.Execute()
	os.Exit(code)
}

func TestTest_Human_SimplePass(t *testing.T) {
	ws := t.TempDir()
	gomod := "module example.com/ami-test\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}
	testSrc := `package pkg
import "testing"
func TestHello(t *testing.T){ if 2!=1+1 { t.Fatal("bad math") } }
`
	if err := os.WriteFile(filepath.Join(ws, "pkg", "a_test.go"), []byte(testSrc), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}
	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestHuman")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_HUMAN=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected non-zero exit code: %v; stdout=\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "test: summary:") && !strings.Contains(string(out), "test PASS") {
		t.Fatalf("expected human summary or pass lines in output; got:\n%s", string(out))
	}
}
