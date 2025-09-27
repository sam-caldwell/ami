package manifest

import (
    "os"
    "path/filepath"
    "testing"
)

func TestIO_BasenamePair_SaveThenLoad(t *testing.T) {
    dir := filepath.Join("build", "test", "manifest", "io_pair")
    _ = os.MkdirAll(dir, 0o755)
    path := filepath.Join(dir, "ami.manifest")
    m := Manifest{Schema: "ami.manifest/v1", Data: map[string]any{"k":"v"}}
    if err := m.Save(path); err != nil { t.Fatalf("save: %v", err) }
    var got Manifest
    if err := got.Load(path); err != nil { t.Fatalf("load: %v", err) }
    if got.Schema != m.Schema { t.Fatalf("schema mismatch") }
}

