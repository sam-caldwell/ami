package mod

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateFromWorkspace_RemoteLatest_NetworkError(t *testing.T) {
	// HOME for cache
	t.Setenv("HOME", t.TempDir())
	// Workspace with a remote import (github.com/...)
	ws := t.TempDir()
	content := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - invalid.invalid/org/repo ==latest
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	err := UpdateFromWorkspace(filepath.Join(ws, "ami.workspace"))
	if err == nil || !errors.Is(err, ErrNetwork) {
		t.Fatalf("expected ErrNetwork from remote resolution; got %v", err)
	}
}
