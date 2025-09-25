package root_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Reuse TestHelper_AmiTestJSON from json test file

func TestTest_JSON_FailSetsExitCode1(t *testing.T) {
	ws := t.TempDir()
	gomod := "module example.com/ami-test\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "a_test.go"), []byte(`package main
import "testing"
func TestFail(t *testing.T){ t.Fatal("nope") }
`), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit code for failing tests; stdout=\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 1 {
			t.Fatalf("unexpected exit code: got %d want 1", code)
		}
	} else {
		t.Fatalf("unexpected error type: %T", err)
	}
}
