package root_test

import (
    "os"
    "path/filepath"
    "testing"
    "os/exec"
)

// Reuse helper TestHelper_AmiBuild

func TestBuild_WritesObjOutputs_NonVerbose(t *testing.T) {
    ws := t.TempDir()
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
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte("package main\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuild")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI=1")
    cmd.Dir = ws
    if err := cmd.Run(); err != nil { t.Fatalf("unexpected non-zero exit: %v", err) }

    out := filepath.Join(ws, "build", "obj", "main", "main.ami.s")
    b, err := os.ReadFile(out)
    if err != nil { t.Fatalf("missing non-debug obj output: %v", err) }
    if len(b) == 0 { t.Fatalf("expected non-empty obj output at %s", out) }
}

