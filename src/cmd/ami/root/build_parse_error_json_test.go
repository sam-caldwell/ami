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

// Helper to run ami build in JSON mode within this process
func TestHelper_AmiBuildJSON_Parse(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_BUILD_JSON_PARSE") != "1" {
		return
	}
	os.Args = []string{"ami", "--json", "build"}
	code := rootcmd.Execute()
	os.Exit(code)
}

// Minimal diag record probe for parse errors
type diagRec struct {
	Schema string `json:"schema"`
	Code   string `json:"code"`
}

func TestBuild_JSON_ParseError_EmitsDiagAndExit1(t *testing.T) {
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
	// invalid AMI source (bad package name)
	bad := "package 123\n"
	if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(bad), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON_Parse")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_BUILD_JSON_PARSE=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for parse error; stdout=\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 1 {
			t.Fatalf("unexpected exit code: got %d want 1; stdout=\n%s", code, string(out))
		}
	} else {
		t.Fatalf("unexpected error type: %T", err)
	}
	// Ensure we saw a diag.v1 with code E_PARSE
	var sawDiag bool
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		var r diagRec
		if json.Unmarshal([]byte(sc.Text()), &r) != nil {
			continue
		}
		if r.Schema == "diag.v1" && (r.Code == "E_PARSE" || r.Code == "E_BAD_PACKAGE" || r.Code == "E_BAD_PACKAGE_BLANK") {
			sawDiag = true
			break
		}
	}
	if !sawDiag {
		t.Fatalf("expected parse-related diag in JSON; stdout=\n%s", string(out))
	}
}
