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

func TestTest_JSON_TimeoutFlagCausesSysIO(t *testing.T) {
	ws := t.TempDir()
	gomod := "module example.com/ami-test\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	slow := `package main
import ("testing"; "time")
func TestSlow(t *testing.T){ time.Sleep(200*time.Millisecond) }
`
	if err := os.WriteFile(filepath.Join(ws, "a_test.go"), []byte(slow), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON_Timeout50ms")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON_TIMEOUT=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit (2) on timeout; stdout=\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 2 {
			t.Fatalf("unexpected exit code: got %d want 2", code)
		}
	} else {
		t.Fatalf("unexpected error type: %T", err)
	}

	// Expect run_start and run_end at minimum
	var sawStart, sawEnd bool
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		var rec struct{ Schema, Type string }
		if json.Unmarshal([]byte(sc.Text()), &rec) != nil {
			continue
		}
		if rec.Schema == "test.v1" && rec.Type == "run_start" {
			sawStart = true
		}
		if rec.Schema == "test.v1" && rec.Type == "run_end" {
			sawEnd = true
		}
	}
	if !sawStart || !sawEnd {
		t.Fatalf("expected test.v1 run_start/run_end events; got:\n%s", string(out))
	}
}
