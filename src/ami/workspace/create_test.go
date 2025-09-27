package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestWorkspace_Create_WritesDefault(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace", "create")
    _ = os.MkdirAll(dir, 0o755)
    path := filepath.Join(dir, "ami.workspace")
    var w Workspace
    if err := w.Create(path); err != nil { t.Fatalf("Create: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil || len(b) == 0 { t.Fatalf("read: %v len=%d", err, len(b)) }
}

