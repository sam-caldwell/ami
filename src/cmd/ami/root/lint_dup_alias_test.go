package root_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/token"
	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestLint_JSON_DuplicateImportAlias(t *testing.T) {
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
	// duplicate alias 'u' refers to two different imports
	src := "package main\nimport u \"example.com/alpha\"\nimport u \"example.com/beta\"\n"
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Capture stdout from lint JSON
	old := os.Args
	// inline capture
	outStr := func(fn func()) string {
		r, w, _ := os.Pipe()
		oldStdout := os.Stdout
		os.Stdout = w
		fn()
		_ = w.Close()
		os.Stdout = oldStdout
		var b strings.Builder
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			b.WriteString(sc.Text())
			b.WriteByte('\n')
		}
		return b.String()
	}
	s := outStr(func() { os.Args = []string{"ami", "--json", "lint"}; _ = rootcmd.Execute() })
	os.Args = old

	// Look for W_DUP_IMPORT_ALIAS
	var seen bool
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		var obj map[string]any
		if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
			continue
		}
		if obj["schema"] == "diag.v1" && obj["code"] == "W_DUP_IMPORT_ALIAS" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("expected W_DUP_IMPORT_ALIAS in lint output; got:\n%s", s)
	}
}
