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

// Helper that runs ami --json lint; enabled via GO_WANT_HELPER_AMI_LINT_JSON=1
func TestHelper_AmiLintJSON(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_LINT_JSON") != "1" && os.Getenv("GO_WANT_HELPER_AMI_JSON") != "1" {
		return
	}
	// Run the CLI in JSON mode
	os.Args = []string{"ami", "--json", "lint"}
	code := rootcmd.Execute()
	os.Exit(code)
}

// Backward-compatible alias: some tests set GO_WANT_HELPER_AMI_JSON
func TestHelper_AmiLintJSON_Alias(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_JSON") != "1" {
		return
	}
	os.Args = []string{"ami", "--json", "lint"}
	code := rootcmd.Execute()
	os.Exit(code)
}

// Helper to run ami --json lint with optional --rules/--max-warn/--strict flags from env.
// AMI_LINT_RULES, AMI_LINT_MAXWARN, AMI_LINT_STRICT ("1" enables)
func TestHelper_AmiLintJSONWithRules(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_LINT_WITH_RULES") != "1" {
		return
	}
	args := []string{"ami", "--json", "lint"}
	if v := strings.TrimSpace(os.Getenv("AMI_LINT_RULES")); v != "" {
		args = append(args, "--rules", v)
	}
	if v := strings.TrimSpace(os.Getenv("AMI_LINT_MAXWARN")); v != "" {
		args = append(args, "--max-warn", v)
	}
	if os.Getenv("AMI_LINT_STRICT") == "1" {
		args = append(args, "--strict")
	}
	os.Args = args
	code := rootcmd.Execute()
	os.Exit(code)
}

// Human-mode helper for lint
func TestHelper_AmiLintHuman(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_LINT_HUMAN") != "1" {
		return
	}
	os.Args = []string{"ami", "lint"}
	code := rootcmd.Execute()
	os.Exit(code)
}

func TestLint_WorkspaceJSONDiagnostics_OnSchemaViolation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ws := t.TempDir()
	// invalid target: absolute path should be rejected by workspace validation
	wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: 1
    target: /abs
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

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSON")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_LINT_JSON=1", "HOME="+home)
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	// For lint, workspace errors are reported but exit code remains 0 (success)
	if err != nil {
		t.Fatalf("unexpected non-zero exit: %v\nstdout/stderr:\n%s", err, string(out))
	}

	// Expect a diag.v1 JSON line with our schema error
	type diag struct {
		Schema  string `json:"schema"`
		Level   string `json:"level"`
		Code    string `json:"code"`
		Message string `json:"message"`
		File    string `json:"file"`
	}
	var seen bool
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var d diag
		if json.Unmarshal([]byte(line), &d) != nil {
			continue
		}
		if d.Schema == "diag.v1" && d.Level == "error" && d.Code == "E_WS_SCHEMA" && strings.Contains(d.Message, "workspace validation failed") {
			if !strings.HasSuffix(d.File, "ami.workspace") {
				t.Fatalf("diag file not set to ami.workspace: %q", d.File)
			}
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("did not observe diag.v1 JSON for workspace schema error. stdout=\n%s", string(out))
	}
}
