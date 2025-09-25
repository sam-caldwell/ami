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

func TestLint_Config_Suppress_PackageRule(t *testing.T) {
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
  linter:
    suppress:
      package:
        main: ["W_UNUSED_IMPORT"]
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
	if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll("src", 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	// Unused import normally triggers W_UNUSED_IMPORT; suppressed via config
	src := "package main\nimport \"fmt\"\n"
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	old := os.Args
	os.Args = []string{"ami", "--json", "lint"}
	out := captureStdoutLint(t, func() { _ = rootcmd.Execute() })
	os.Args = old

	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var obj map[string]any
		if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
			continue
		}
		if obj["schema"] == "diag.v1" && obj["code"] == "W_UNUSED_IMPORT" {
			t.Fatalf("unexpected W_UNUSED_IMPORT (should be suppressed):\n%s", out)
		}
	}
}

func TestLint_Config_Suppress_PathRule(t *testing.T) {
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
  linter:
    suppress:
      paths:
        ./src/**: ["W_PKG_LOWERCASE"]
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
	if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll("src", 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	// Uppercase package normally triggers W_PKG_LOWERCASE; suppressed by path glob
	src := "package Main\n"
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	old := os.Args
	os.Args = []string{"ami", "--json", "lint"}
	out := captureStdoutLint(t, func() { _ = rootcmd.Execute() })
	os.Args = old

	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var obj map[string]any
		if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
			continue
		}
		if obj["schema"] == "diag.v1" && obj["code"] == "W_PKG_LOWERCASE" {
			t.Fatalf("unexpected W_PKG_LOWERCASE (should be suppressed by path):\n%s", out)
		}
	}
}
