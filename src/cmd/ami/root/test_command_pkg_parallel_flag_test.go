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

// Helpers invoked via environment already defined elsewhere.

func TestTest_JSON_PkgParallelFlagPropagatesToRunStart(t *testing.T) {
	ws := t.TempDir()
	gomod := "module example.com/ami-test\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// two packages
	if err := os.MkdirAll(filepath.Join(ws, "p1"), 0o755); err != nil {
		t.Fatalf("mkdir p1: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "p2"), 0o755); err != nil {
		t.Fatalf("mkdir p2: %v", err)
	}
	src := `package p
import "testing"
func TestX(t *testing.T){ }
`
	if err := os.WriteFile(filepath.Join(ws, "p1", "a_test.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write p1 test: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "p2", "a_test.go"), []byte(strings.Replace(src, "package p", "package q", 1)), 0o644); err != nil {
		t.Fatalf("write p2 test: %v", err)
	}

	// Run with --pkg-parallel=2 (via helper wrapper)
	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON_ParallelPkg2")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON_PKGPAR=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected non-zero exit: %v, stdout=\n%s", err, string(out))
	}
	// Examine run_start for pkg_parallel=2
	type runStart struct {
		Schema, Type string
		PkgParallel  int `json:"pkg_parallel"`
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		var rs runStart
		if json.Unmarshal([]byte(sc.Text()), &rs) != nil {
			continue
		}
		if rs.Schema == "test.v1" && rs.Type == "run_start" {
			if rs.PkgParallel != 2 {
				t.Fatalf("expected pkg_parallel=2; got %d", rs.PkgParallel)
			}
			return
		}
	}
	t.Fatalf("did not find run_start with pkg_parallel=2; stdout=\n%s", string(out))
}
