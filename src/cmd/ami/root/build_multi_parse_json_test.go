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

func TestBuild_JSON_MultipleParserDiagnostics_Exit1(t *testing.T) {
	ws := t.TempDir()
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
	// two invalid units
	if err := os.WriteFile(filepath.Join(ws, "src", "a.ami"), []byte("package 123\n"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "src", "b.ami"), []byte("package _\n"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit 1; stdout=\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 1 {
			t.Fatalf("got exit %d want 1; stdout=\n%s", code, string(out))
		}
	} else {
		t.Fatalf("unexpected err type: %T", err)
	}

	// Count diag.v1 lines for parse-related codes
	count := 0
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		var r map[string]any
		if json.Unmarshal([]byte(sc.Text()), &r) != nil {
			continue
		}
		if r["schema"] == "diag.v1" {
			if code, _ := r["code"].(string); code == "E_PARSE" || strings.HasPrefix(code, "E_BAD_PACKAGE") {
				count++
			}
		}
	}
	if count < 2 {
		t.Fatalf("expected multiple parser diagnostics; got %d; stdout=\n%s", count, string(out))
	}
}
