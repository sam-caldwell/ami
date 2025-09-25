package manifest

import (
    "os"
    "path/filepath"
    "testing"
)

func TestManifest_SaveLoad(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "ami.manifest")
    m := &Manifest{Project: Project{Name:"demo", Version:"0.0.1"}}
    if err := Save(path, m); err != nil { t.Fatalf("save: %v", err) }
    got, err := Load(path)
    if err != nil { t.Fatalf("load: %v", err) }
    if got.Project.Name != "demo" { t.Fatalf("unexpected project: %+v", got.Project) }
    if _, err := os.Stat(path); err != nil { t.Fatalf("stat: %v", err) }
}
