package workspace

import (
    "path/filepath"
    "testing"
)

func TestManifest_SaveAndLoad_Empty(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "ami.sum")
    var m Manifest
    if err := m.Save(path); err != nil { t.Fatalf("Save: %v", err) }
    var m2 Manifest
    if err := m2.Load(path); err != nil { t.Fatalf("Load: %v", err) }
}

