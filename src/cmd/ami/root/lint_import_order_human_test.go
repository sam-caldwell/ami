package root_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLint_Human_ImportOrder_Warns(t *testing.T) {
	ws := t.TempDir()
	wsContent := `version: 1.0.0
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
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := "package main\nimport \"b/zz\"\nimport \"a/aa\"\n"
	if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	// human lint helper
	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintHuman")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_LINT_HUMAN=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected non-zero: %v; out=\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "W_IMPORT_ORDER: imports are not ordered") {
		t.Fatalf("expected W_IMPORT_ORDER in human output; got:\n%s", string(out))
	}
}
