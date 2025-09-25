package root_test

import (
	"os"
	"path/filepath"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

func TestBuild_NoDebugArtifacts_WhenNotVerbose(t *testing.T) {
	ws := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(ws); err != nil {
		t.Fatalf("chdir: %v", err)
	}

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
	if err := os.WriteFile("ami.workspace", []byte(wsContent), 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}

	// run build (no --verbose)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "build"}
	_ = rootcmd.Execute()

	if _, err := os.Stat(filepath.Join("build", "debug")); !os.IsNotExist(err) {
		t.Fatalf("expected no build/debug directory; err=%v", err)
	}
	if _, err := os.Stat("ami.manifest"); err != nil {
		t.Fatalf("expected ami.manifest to be written, err=%v", err)
	}
}
