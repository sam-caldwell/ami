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

// capture helper shared with lint_entrypoints_test.go if needed
func captureStdoutLintRules(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var b strings.Builder
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		b.WriteString(sc.Text())
		b.WriteByte('\n')
	}
	return b.String()
}

func TestLint_JSON_BasicRules(t *testing.T) {
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
  linter: {}
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
	// CRLF and no trailing LF, uppercase package, unused import
	src := "package Main\r\nimport \"fmt\"" // no newline at end
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Snapshot and restore os.Args around Execute
	oldArgs := os.Args
	out := captureStdoutLintRules(t, func() {
		os.Args = []string{"ami", "--json", "lint"}
		_ = rootcmd.Execute()
	})
	os.Args = oldArgs

	// Expect warnings: W_PKG_LOWERCASE, W_UNUSED_IMPORT, W_FILE_NO_NEWLINE, W_FILE_CRLF
	want := map[string]bool{"W_PKG_LOWERCASE": false, "W_UNUSED_IMPORT": false, "W_FILE_NO_NEWLINE": false, "W_FILE_CRLF": false}
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
		if obj["schema"] == "diag.v1" {
			if code, _ := obj["code"].(string); want[code] == false {
				want[code] = true
			}
		}
	}
	for k, v := range want {
		if !v {
			t.Fatalf("expected lint code %s in output; got:\n%s", k, out)
		}
	}
}

func TestLint_JSON_DuplicateImport(t *testing.T) {
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
  linter: {}
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
	// duplicate imports
	src := "package main\nimport \"fmt\"\nimport \"fmt\"\n"
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	oldArgs2 := os.Args
	out := captureStdoutLintRules(t, func() {
		os.Args = []string{"ami", "--json", "lint"}
		_ = rootcmd.Execute()
	})
	os.Args = oldArgs2

	var seen bool
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var obj map[string]any
		if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
			continue
		}
		if obj["schema"] == "diag.v1" && obj["code"] == "W_DUP_IMPORT" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("expected W_DUP_IMPORT in lint output; got:\n%s", out)
	}
}
