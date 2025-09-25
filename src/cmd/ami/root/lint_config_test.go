package root_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

// Verifies workspace-configured severities and in-file pragma suppression.
func TestLint_Config_SeverityAndPragmaSuppress(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	_, restore := testutil.ChdirToBuildTest(t)
	defer restore()

	ws := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env: []
  linker: {}
  linter:
    rules:
      W_FILE_NO_NEWLINE: off
      W_PKG_LOWERCASE: error
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
	if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll("src", 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	// Source triggers: uppercase package (W_PKG_LOWERCASE->error), unused import (suppressed by pragma),
	// no trailing newline (off via config)
	src := "#pragma lint:disable W_UNUSED_IMPORT\npackage Main\nimport \"fmt\"" // no newline at end
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Run CLI in JSON mode
	old := os.Args
	os.Args = []string{"ami", "--json", "lint"}
	out := captureStdoutLint(t, func() { _ = rootcmd.Execute() })
	os.Args = old

	// Expectations:
	// - W_UNUSED_IMPORT must NOT appear (pragma disabled)
	// - W_FILE_NO_NEWLINE must NOT appear (off in config)
	// - W_PKG_LOWERCASE must appear with level "error"
	var seenLowerAsError bool
	var seenUnused bool
	var seenNoNewline bool
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var obj map[string]any
		if json.Unmarshal([]byte(line), &obj) != nil {
			continue
		}
		if obj["schema"] != "diag.v1" {
			continue
		}
		code, _ := obj["code"].(string)
		switch code {
		case "W_UNUSED_IMPORT":
			seenUnused = true
		case "W_FILE_NO_NEWLINE":
			seenNoNewline = true
		case "W_PKG_LOWERCASE":
			if lvl, _ := obj["level"].(string); lvl == "error" {
				seenLowerAsError = true
			}
		}
	}
	if seenUnused {
		t.Fatalf("unexpected W_UNUSED_IMPORT in output (should be suppressed).\n%s", out)
	}
	if seenNoNewline {
		t.Fatalf("unexpected W_FILE_NO_NEWLINE in output (should be off).\n%s", out)
	}
	if !seenLowerAsError {
		t.Fatalf("expected W_PKG_LOWERCASE with level error.\n%s", out)
	}
}
