package mod

import (
    "path/filepath"
    "testing"
)

func TestFindWorkspaceRoot_ErrorWhenNotFound(t *testing.T) {
    dir := t.TempDir()
    nested := filepath.Join(dir, "a", "b")
    if root, err := findWorkspaceRoot(nested); err == nil || root != "" {
        t.Fatalf("expected error when ami.workspace not found; got %q, err=%v", root, err)
    }
}

