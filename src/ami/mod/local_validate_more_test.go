package mod

import (
    "os"
    "path/filepath"
    "testing"
)

func TestFindWorkspaceRoot_WalksUp(t *testing.T) {
    ws := t.TempDir()
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\n"), 0o644); err != nil {
        t.Fatalf("write ws: %v", err)
    }
    nested := filepath.Join(ws, "a", "b", "c")
    if err := os.MkdirAll(nested, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    got, err := findWorkspaceRoot(nested)
    if err != nil { t.Fatalf("findWorkspaceRoot: %v", err) }
    if filepath.Clean(got) != filepath.Clean(ws) {
        t.Fatalf("got %s want %s", got, ws)
    }
}

