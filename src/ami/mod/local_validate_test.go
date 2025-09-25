package mod

import (
    "os"
    "path/filepath"
    "testing"
)

func TestIsDeclaredLocalImport(t *testing.T) {
    ws := t.TempDir()
    content := `version: 1.0.0
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
      import:
        - ./subproject ==latest
        - ./other v1.0.0
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    ok, err := isDeclaredLocalImport(filepath.Join(ws, "ami.workspace"), "./subproject")
    if err != nil { t.Fatalf("err: %v", err) }
    if !ok { t.Fatalf("expected declared import") }
    ok, err = isDeclaredLocalImport(filepath.Join(ws, "ami.workspace"), "./missing")
    if err != nil { t.Fatalf("err: %v", err) }
    if ok { t.Fatalf("did not expect declared import") }
}

func TestGet_Local_WithinWorkspaceAndDeclared(t *testing.T) {
    // HOME for cache
    home := t.TempDir()
    t.Setenv("HOME", home)
    // Workspace
    ws := t.TempDir()
    // Subproject
    sub := filepath.Join(ws, "subproject")
    if err := os.MkdirAll(sub, 0o755); err != nil { t.Fatalf("mkdir sub: %v", err) }
    if err := os.WriteFile(filepath.Join(sub, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // ami.workspace with import declaration
    content := `version: 1.0.0
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
      import:
        - ./subproject ==latest
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // Run from workspace
    old, _ := os.Getwd()
    defer func(){ _ = os.Chdir(old) }()
    if err := os.Chdir(ws); err != nil { t.Fatalf("chdir: %v", err) }
    dest, err := Get("./subproject")
    if err != nil { t.Fatalf("Get local: %v", err) }
    if fi, err := os.Stat(dest); err != nil || !fi.IsDir() { t.Fatalf("dest missing: %v", err) }
}

func TestGet_Local_RequiresDeclaration(t *testing.T) {
    // HOME for cache
    home := t.TempDir()
    t.Setenv("HOME", home)
    // Workspace
    ws := t.TempDir()
    // Subproject exists but not declared
    sub := filepath.Join(ws, "subproject")
    if err := os.MkdirAll(sub, 0o755); err != nil { t.Fatalf("mkdir sub: %v", err) }
    if err := os.WriteFile(filepath.Join(sub, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // ami.workspace without import entry
    content := `version: 1.0.0
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
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // Run from workspace
    old, _ := os.Getwd()
    defer func(){ _ = os.Chdir(old) }()
    if err := os.Chdir(ws); err != nil { t.Fatalf("chdir: %v", err) }
    if _, err := Get("./subproject"); err == nil {
        t.Fatalf("expected error when local path not declared in workspace")
    }
}

