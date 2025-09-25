package manifest

import (
    "os"
    "path/filepath"
    "testing"
)

func TestManifest_Save_DeterministicOrdering(t *testing.T) {
    dir := t.TempDir()
    path1 := filepath.Join(dir, "a1.manifest")
    path2 := filepath.Join(dir, "a2.manifest")

    m1 := &Manifest{
        Project: Project{Name: "demo", Version: "0.0.1"},
        Packages: []Package{{Name:"b", Version:"v1.0.0"}, {Name:"a", Version:"v1.0.0"}},
        Artifacts: []Artifact{{Path:"/z"}, {Path:"/a"}},
    }
    if err := Save(path1, m1); err != nil { t.Fatalf("save1: %v", err) }

    // Same items but different input order; output should be identical
    m2 := &Manifest{
        Project: Project{Name: "demo", Version: "0.0.1"},
        Packages: []Package{{Name:"a", Version:"v1.0.0"}, {Name:"b", Version:"v1.0.0"}},
        Artifacts: []Artifact{{Path:"/a"}, {Path:"/z"}},
    }
    if err := Save(path2, m2); err != nil { t.Fatalf("save2: %v", err) }

    b1, _ := os.ReadFile(path1)
    b2, _ := os.ReadFile(path2)
    if string(b1) != string(b2) {
        t.Fatalf("manifest writes are not deterministic:\n%s\n---\n%s", string(b1), string(b2))
    }
}

