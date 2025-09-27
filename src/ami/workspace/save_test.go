package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestWorkspace_Save_PrefixesYAMLDocMarker(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace", "save")
    _ = os.MkdirAll(dir, 0o755)
    path := filepath.Join(dir, "ami.workspace")
    w := DefaultWorkspace()
    if err := w.Save(path); err != nil { t.Fatalf("Save: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    if len(b) < 3 || string(b[:4]) != "---\n" { t.Fatalf("expected YAML doc marker at start, got: %q", string(b[:4])) }
}

