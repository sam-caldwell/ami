package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileBackend_WithFileScheme(t *testing.T) {
	// HOME for cache
	t.Setenv("HOME", t.TempDir())
	// Workspace and subproject
	ws := t.TempDir()
	sub := filepath.Join(ws, "subproject")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	content := `version: 1.0.0
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - ./subproject ==latest
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(ws)

	dest, pkg, ver, err := GetWithInfo("file://./subproject")
	if err != nil {
		t.Fatalf("GetWithInfo file://: %v", err)
	}
	if pkg != "" || ver != "" {
		t.Fatalf("file backend should not set pkg/ver; got %q %q", pkg, ver)
	}
	if fi, err := os.Stat(dest); err != nil || !fi.IsDir() {
		t.Fatalf("dest missing: %v", err)
	}
}
