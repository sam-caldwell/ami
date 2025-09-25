package mod

import (
    "os"
    "path/filepath"
    "testing"
)

func TestUpdateFromWorkspace_InvalidConstraint_Error(t *testing.T) {
    // HOME for cache
    t.Setenv("HOME", t.TempDir())
    // Workspace with unsupported operator "<="
    ws := t.TempDir()
    content := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - github.com/example/repo <=v1.0.0
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := UpdateFromWorkspace(filepath.Join(ws, "ami.workspace")); err == nil {
        t.Fatalf("expected error for unsupported operator")
    }
}

