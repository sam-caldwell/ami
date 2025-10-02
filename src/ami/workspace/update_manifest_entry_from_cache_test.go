package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestUpdateManifestEntryFromCache_EmptyDir(t *testing.T) {
    dir := t.TempDir()
    // create package/version dir and file
    p := filepath.Join(dir, "name", "1.0.0")
    if err := os.MkdirAll(p, 0o755); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(p, "a.txt"), []byte("x"), 0o644); err != nil { t.Fatal(err) }
    var m Manifest
    if err := UpdateManifestEntryFromCache(&m, dir, "name", "1.0.0"); err != nil { t.Fatalf("err: %v", err) }
    if !m.Has("name", "1.0.0") { t.Fatalf("expected manifest entry") }
}

