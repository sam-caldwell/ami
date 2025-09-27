package manifest

import (
    "os"
    "path/filepath"
    "testing"
)

func TestManifest_Save_FailsWhenPathIsDirectory(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_manifest", "save_dir")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create a directory that collides with the file path
    target := filepath.Join(dir, "manifest_dir")
    if err := os.MkdirAll(target, 0o755); err != nil { t.Fatalf("mkdir target: %v", err) }
    m := Manifest{Schema: "ami.manifest/v1", Data: map[string]any{"k":"v"}}
    if err := m.Save(target); err == nil { t.Fatalf("expected error when saving to a directory path") }
}

