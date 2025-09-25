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

func TestLint_JSON_Positions_AttachedWhereAvailable(t *testing.T) {
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
	// Construct a file with:
	// - Uppercase package (W_PKG_LOWERCASE) at line 1
	// - CRLF (W_FILE_CRLF)
	// - Duplicate import (W_DUP_IMPORT)
	// - No trailing newline (W_FILE_NO_NEWLINE)
	src := "package Main\r\nimport \"fmt\"\nimport \"fmt\"" // no trailing \n
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	old := os.Args
	os.Args = []string{"ami", "--json", "lint"}
	out := captureStdoutLint(t, func() { _ = rootcmd.Execute() })
	os.Args = old

	// Expect each of these codes to include a non-nil pos with integer fields
	want := map[string]bool{"W_PKG_LOWERCASE": false, "W_DUP_IMPORT": false, "W_FILE_NO_NEWLINE": false, "W_FILE_CRLF": false}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var obj map[string]any
		if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
			continue
		}
		if obj["schema"] != "diag.v1" {
			continue
		}
		code, _ := obj["code"].(string)
		pos, _ := obj["pos"].(map[string]any)
		if _, ok := want[code]; ok {
			if pos != nil {
				// basic sanity: positive line/column
				if l, lok := pos["line"].(float64); lok && l >= 1 {
					if c, cok := pos["column"].(float64); cok && c >= 1 {
						want[code] = true
					}
				}
			}
		}
	}
	for k, v := range want {
		if !v {
			t.Fatalf("expected position for %s in output; got:\n%s", k, out)
		}
	}
}
