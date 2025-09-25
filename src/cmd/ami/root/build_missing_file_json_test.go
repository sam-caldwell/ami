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

// Reuse TestHelper_AmiBuildJSON from build_integrity_logs_test.go

func TestBuild_JSON_MissingFile_SystemIOError(t *testing.T) {
	ws := t.TempDir()
	// minimal workspace
	wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env: []
  linker: {}
  linter: {}
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	// create source then remove read permission to trigger read error
	srcPath := filepath.Join(ws, "src", "main.ami")
	if err := os.WriteFile(srcPath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	if err := os.Chmod(srcPath, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for system I/O error; stdout=\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 2 {
			t.Fatalf("unexpected exit code: got %d want 2; stdout=\n%s", code, string(out))
		}
	} else {
		t.Fatalf("unexpected error type: %T", err)
	}

	// Confirm a diag.v1 line was emitted
	var saw bool
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		var r map[string]any
		if json.Unmarshal([]byte(sc.Text()), &r) != nil {
			continue
		}
		if r["schema"] == "diag.v1" {
			saw = true
			break
		}
	}
	if !saw {
		t.Fatalf("expected diag.v1 JSON output; stdout=\n%s", string(out))
	}
}
