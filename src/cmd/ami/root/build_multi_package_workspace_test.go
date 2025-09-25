package root_test

import (
    "os"
    "path/filepath"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

// Ensures build compiles all workspace packages and writes deterministic layout.
func TestBuild_WorkspaceMultiPackage_DebugArtifactsExist(t *testing.T) {
    t.Setenv("AMI_SEM_DIAGS", "0")
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    ws := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
  - util: { version: 0.0.1, root: ./lib, import: [] }
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.MkdirAll("lib", 0o755); err != nil { t.Fatalf("mkdir lib: %v", err) }
    if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte("package main\nimport \"util\"\n"), 0o644); err != nil { t.Fatalf("write main: %v", err) }
    if err := os.WriteFile(filepath.Join("lib", "lib.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatalf("write lib: %v", err) }

    // Run build --verbose (writes debug artifacts and obj outputs)
    oldArgs := os.Args
    defer func(){ os.Args = oldArgs }()
    os.Args = []string{"ami", "build", "--verbose"}
    _ = rootcmd.Execute()

    // Expect AST artifacts for both packages under deterministic paths
    if _, err := os.Stat(filepath.Join("build","debug","ast","main","main.ami.ast.json")); err != nil { t.Fatalf("missing main ast: %v", err) }
    if _, err := os.Stat(filepath.Join("build","debug","ast","util","lib.ami.ast.json")); err != nil { t.Fatalf("missing util ast: %v", err) }

    // Expect obj outputs for both packages
    if _, err := os.Stat(filepath.Join("build","obj","main","index.json")); err != nil { t.Fatalf("missing main obj index: %v", err) }
    if _, err := os.Stat(filepath.Join("build","obj","util","index.json")); err != nil { t.Fatalf("missing util obj index: %v", err) }
}

