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

func TestLint_Hints_Positions_Attached(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	_, restore := testutil.ChdirToBuildTest(t)
	defer restore()

	ws := `version: 1.0.0
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
	if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll("src", 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	// Contains Owned<string>, map<slice<byte>,int>, and *Node
	src := "package p\nstruct S { R Owned<string>; M map<slice<byte>, int>; N *Node }\n"
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	old := os.Args
	os.Args = []string{"ami", "--json", "lint"}
	out := captureStdoutLint(t, func() { _ = rootcmd.Execute() })
	os.Args = old

	want := map[string]bool{"W_RAII_OWNED_HINT": false, "W_MAP_KEY_TYPE_HINT": false}
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
		if _, ok := want[code]; !ok {
			continue
		}
		if pos, ok := obj["pos"].(map[string]any); ok {
			if l, lok := pos["line"].(float64); lok && l >= 1 {
				want[code] = true
			}
		}
	}
	for k, v := range want {
		if !v {
			t.Fatalf("expected position for %s; out=\n%s", k, out)
		}
	}
}
