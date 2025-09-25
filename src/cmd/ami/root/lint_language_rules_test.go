package root_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// Helper to capture stdout of a function
func captureStdoutLang(t *testing.T, fn func()) string {
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

func TestLint_JSON_Emits_W_LANG_NOT_GO_And_W_GO_SYNTAX_DETECTED(t *testing.T) {
	ws := t.TempDir()
	// Minimal workspace
	content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	// Trigger both: uppercase package name (any unit will emit W_LANG_NOT_GO), and Go-like syntax (:=)
	src := "package Main\nvar x int\nx := 1\n"
	if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	// Run `ami --json lint`
	oldDir, _ := os.Getwd()
	_ = os.Chdir(ws)
	defer os.Chdir(oldDir)
	oldArgs := os.Args
	out := captureStdoutLang(t, func() { os.Args = []string{"ami", "--json", "lint"}; _ = rootcmd.Execute() })
	os.Args = oldArgs
	// Scan for codes
	seenLang := false
	seenGo := false
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var m map[string]any
		if json.Unmarshal([]byte(sc.Text()), &m) != nil {
			continue
		}
		if m["schema"] != "diag.v1" {
			continue
		}
		switch m["code"] {
		case "W_LANG_NOT_GO":
			seenLang = true
		case "W_GO_SYNTAX_DETECTED":
			seenGo = true
		}
	}
	if !seenLang || !seenGo {
		t.Fatalf("expected both W_LANG_NOT_GO and W_GO_SYNTAX_DETECTED; out=\n%s", out)
	}
}
