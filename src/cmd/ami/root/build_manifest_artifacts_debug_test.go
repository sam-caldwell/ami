package root_test

import (
	"os"
	"path/filepath"
	"testing"

	man "github.com/sam-caldwell/ami/src/ami/manifest"
	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestBuild_Manifest_IncludesAllDebugArtifactsAcrossPackages(t *testing.T) {
	t.Setenv("AMI_SEM_DIAGS", "0")
	_, restore := testutil.ChdirToBuildTest(t)
	defer restore()

	ws := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main: { version: 0.0.1, root: ./src, import: [util] }
  - util: { version: 0.0.1, root: ./lib, import: [] }
`
	if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll("src", 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.MkdirAll("lib", 0o755); err != nil {
		t.Fatalf("mkdir lib: %v", err)
	}
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte("package main\nimport \"util\"\n"), 0o644); err != nil {
		t.Fatalf("write main: %v", err)
	}
	if err := os.WriteFile(filepath.Join("lib", "lib.ami"), []byte("package util\n"), 0o644); err != nil {
		t.Fatalf("write lib: %v", err)
	}

	// Build verbose to emit debug artifacts
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "build", "--verbose"}
	_ = rootcmd.Execute()

	m, err := man.Load("ami.manifest")
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	have := map[string]bool{}
	for _, a := range m.Artifacts {
		have[a.Path] = true
	}

	// Required debug artifacts across packages
	must := []string{
		filepath.Join("build", "debug", "source", "resolved.json"),
		filepath.Join("build", "debug", "ast", "main", "main.ami.ast.json"),
		filepath.Join("build", "debug", "ast", "util", "lib.ami.ast.json"),
		filepath.Join("build", "debug", "ir", "main", "main.ami.ir.json"),
		filepath.Join("build", "debug", "ir", "util", "lib.ami.ir.json"),
		filepath.Join("build", "debug", "ir", "main", "main.ami.pipelines.json"),
		filepath.Join("build", "debug", "ir", "util", "lib.ami.pipelines.json"),
		filepath.Join("build", "debug", "ir", "main", "main.ami.eventmeta.json"),
		filepath.Join("build", "debug", "ir", "util", "lib.ami.eventmeta.json"),
		filepath.Join("build", "debug", "asm", "main", "index.json"),
		filepath.Join("build", "debug", "asm", "util", "index.json"),
		filepath.Join("build", "debug", "asm", "main", "main.ami.s"),
	}
	for _, p := range must {
		if !have[p] {
			t.Fatalf("manifest missing artifact %s", p)
		}
	}
}
